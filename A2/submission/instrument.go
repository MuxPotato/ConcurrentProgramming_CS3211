package submission

import (
	"assign2/utils"
	"assign2/wg"
	"context"
	"sort"
)

type Order struct {
	id        uint32
	price     uint32
	count     uint32
	matchId   uint32
	timestamp int64
}

type Instrument struct {
	buyOrders    []Order
	sellOrders   []Order
	BuyOrdersCh  chan Request
	SellOrdersCh chan Request
}

type Request struct {
	order    utils.Input
	outputCh chan<- bool
}

func lessThan(a uint32, b uint32) bool {
	return a < b;
}

func greaterThan(a uint32, b uint32) bool {
	return a > b;
}

func (instrument *Instrument) Init(ctx context.Context, name string, mainWg *wg.WaitGroup) {
	instrument.buyOrders = make([]Order, 0, 10)
	instrument.sellOrders = make([]Order, 0, 10)
	instrument.BuyOrdersCh = make(chan Request, 50)
	instrument.SellOrdersCh = make(chan Request, 50)

	// This is the master goroutine for the instrument, it starts the goroutines responsible for processing orders
	go func() {
		// Engine must wait for all instrument goroutines to terminate first before terminating
		defer mainWg.Done()
		var phaseWait wg.WaitGroup

		// Repeatedly read orders for this instrument and start the buy/sell phase accordingly
		for {
			select {
			case request := <-instrument.BuyOrdersCh:
				// fmt.Fprintf(os.Stderr, "Starting buy phase for %s\n", name)

				// Put the buy order back so that it can be processed
				instrument.BuyOrdersCh <- request
				phaseWait.Add(1)
				handlePhase(ctx, &instrument.buyOrders, &instrument.sellOrders, instrument.BuyOrdersCh, greaterThan, &phaseWait)
				// Wait for all the buy orders to finish processing
				phaseWait.Wait()

				// fmt.Fprintf(os.Stderr, "Ending buy phase for %s, #buyOrders = %d, #sellOrders = %d\n", name, len(instrument.buyOrders), len(instrument.sellOrders))
			case request := <-instrument.SellOrdersCh:
				// fmt.Fprintf(os.Stderr, "Starting sell phase for %s\n", name)

				// Put the sell order back so that it can be processed
				instrument.SellOrdersCh <- request
				phaseWait.Add(1)
				handlePhase(ctx, &instrument.sellOrders, &instrument.buyOrders, instrument.SellOrdersCh, lessThan, &phaseWait)
				// Wait for all the sell orders to finish processing
				phaseWait.Wait()

				// fmt.Fprintf(os.Stderr, "Ending sell phase for %s, #buyOrders = %d, #sellOrders = %d\n", name, len(instrument.buyOrders), len(instrument.sellOrders))
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Match against the first resting order as much as possible, based on the resting and active orders' count
// It is assumed that resting orders have already been sorted best first
func matchOrder(orders *[]Order, inputCount *uint32) (uint32, uint32, uint32, uint32, int64) {
	id := (*orders)[0].id;
	price := (*orders)[0].price;
	count := min((*orders)[0].count, *inputCount);
	match := (*orders)[0].matchId;

	*inputCount -= count;
	if ((*orders)[0].count == count) {
		*orders = (*orders)[1:]
	} else {
		(*orders)[0].count -= count;
		(*orders)[0].matchId++;
	}

	return id, price, count, match, GetCurrentTimestamp()
}

// Look for an order with a specific id to cancel, and return whether such an order was found
func cancelOrder(orders *[]Order, orderId uint32) bool {
	for i, order := range *orders {
		if order.id == orderId {
			*orders = append((*orders)[:i], (*orders)[i + 1:]...)
			return true
		}
	}
	return false
}

func handlePhase(ctx context.Context, sameOrders, oppositeOrders *[]Order, ordersCh <-chan Request, cmp func(uint32, uint32) bool, wg *wg.WaitGroup) {
	matchingCtx, cancel := context.WithCancel(context.Background())
	unmatchedOrdersch := make(chan Request)

	// This goroutine handles matching against resting orders of the opposite side
	go func() {
		// Make sure to shut down the other goroutine after this one is done processing all orders
		defer cancel()

		// Repeatedly read and process orders until there are none left
		for {
			select {
			case request := <-ordersCh:
				if request.order.OrderType == utils.InputCancel {
					// Process a request to cancel an order on the opposite side
					cancelled := cancelOrder(oppositeOrders, request.order.OrderId)
					utils.OutputOrderDeleted(request.order, cancelled, GetCurrentTimestamp())
					request.outputCh <- cancelled
				} else {
					// Match the order as much as possible
					for request.order.Count > 0 && len(*oppositeOrders) > 0 && !cmp((*oppositeOrders)[0].price, request.order.Price) {
						id, price, count, match, timestamp := matchOrder(oppositeOrders, &request.order.Count)
						utils.OutputOrderExecuted(id, request.order.OrderId, match, price, count, timestamp)
					}

					if (request.order.Count > 0) {
						// If the order was not fully matched, pass it to the other goroutine to add to the order book
						unmatchedOrdersch <- request
					} else {
						// If the order was fully matched, inform the client immediately
						request.outputCh <- true
					}
				}
			case <-ctx.Done():
				return
			default:
				// No orders left to process, almost ready to terminate this phase
				return
			}
		}
	}()

	// This goroutine handles storing unmatched orders of this side
	go func() {
		defer wg.Done()

		// Repeatedly read and process orders until there are none left
		for {
			select {
			case request := <-unmatchedOrdersch:
				// Add the unmatched order to the order book
				timestamp := GetCurrentTimestamp()
				*sameOrders = append(*sameOrders, Order{request.order.OrderId, request.order.Price, request.order.Count, 1, timestamp})
				utils.OutputOrderAdded(request.order, timestamp)
				// Inform the client that the order was not fully matched
				request.outputCh <- false
			case <-matchingCtx.Done():
				// All orders for this phase have been processed, sort the added orders before ending the phase
				sort.Slice(*sameOrders, func(i, j int) bool {
					if ((*sameOrders)[i].price == (*sameOrders)[j].price) {
						return (*sameOrders)[i].timestamp < (*sameOrders)[j].timestamp
					}
					return cmp((*sameOrders)[i].price, (*sameOrders)[j].price)
				})
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}
