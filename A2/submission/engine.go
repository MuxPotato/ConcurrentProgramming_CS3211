package submission

import "C"
import (
	"assign2/utils"
	"assign2/wg"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type InstrumentRequest struct {
	instrument string
	outputCh   chan<- Instrument
}

type UnmatchedOrder struct {
    instrument string
    orderType  utils.InputType
}

type Engine struct {
	wg                  *wg.WaitGroup
    instrumentRequestCh chan InstrumentRequest
}

func (e *Engine) Init(ctx context.Context, wg *wg.WaitGroup) {
    wg.Add(1)
	e.wg = wg
    e.instrumentRequestCh = make(chan InstrumentRequest)

    // Goroutine to handle requests for an instrument from clients
    go func() {
        // Engine must wait for this goroutine to terminate first before terminating
        defer wg.Done()
        defer close(e.instrumentRequestCh)
        // Map to keep track of the instruments
        instrumentMap := make(map[string]Instrument)

        // Repeatedly read and process requests for instruments
        for {
            select {
            case request := <-e.instrumentRequestCh:
                // Create instrument if does not exist, then return it
                instrument, ok := instrumentMap[request.instrument]
                if !ok {
                    wg.Add(1)
                    instrument.Init(ctx, request.instrument, wg);
                    instrumentMap[request.instrument] = instrument
                }
                request.outputCh <- instrument
            case <-ctx.Done():
                return
            }
        }
    }()
}

func (e *Engine) Shutdown(ctx context.Context) {
	e.wg.Wait()
}

func (e *Engine) Accept(ctx context.Context, conn net.Conn) {
	e.wg.Add(2)

	go func() {
		defer e.wg.Done()
		<-ctx.Done()
		conn.Close()
	}()

	// This goroutine handles the connection.
	go func() {
		defer e.wg.Done()
		handleConn(conn, e.instrumentRequestCh)
	}()
}

func handleConn(conn net.Conn, instrumentRequestCh chan<- InstrumentRequest) {
	defer conn.Close()
    // Channel for receiving requested instrument
    instrumentResultCh := make(chan Instrument)
    // Channel for receiving order matching result
    matchingResultCh := make(chan bool)
    // Map to keep track of the instrument and type of unmatched and hence, cancellable orders
    unmatchedOrders := make(map[uint32]UnmatchedOrder)

	for {
		in, err := utils.ReadInput(conn)
		if err != nil {
			if err != io.EOF {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			return
		}
        // switch in.OrderType {
        // case utils.InputCancel:
        //     fmt.Fprintf(os.Stderr, "Got cancel ID: %v\n", in.OrderId)
        // default:
        //     fmt.Fprintf(os.Stderr, "Got order: %c %v x %v @ %v ID: %v\n", in.OrderType, in.Instrument, in.Count, in.Price, in.OrderId)
        // }
        switch in.OrderType {
        case utils.InputCancel:
            order, ok := unmatchedOrders[in.OrderId]
            if ok {
                // There may be an unmatched order with this ID, get its instrument
                instrumentRequestCh <- InstrumentRequest{order.instrument, instrumentResultCh}
                instrument := <-instrumentResultCh

                // Send the cancel request to the appropriate channel
                // This is opposite of a matching request, since each order channel operates on the array storing orders of the opposite side
                if order.orderType == utils.InputBuy {
                    instrument.SellOrdersCh <- Request{in, matchingResultCh}
                } else {
                    instrument.BuyOrdersCh <- Request{in, matchingResultCh}
                }

                // Wait till cancel request is done, result doesn't matter since printing is done by the processer
                <-matchingResultCh
                // Delete entry; in Go it is allowed to call delete on keys that don't exist
                delete(unmatchedOrders, in.OrderId)
            } else {
                // There is definitely no unmatched order with this ID, just print cancel rejected
                utils.OutputOrderDeleted(in, false, GetCurrentTimestamp())
            }
        default:
            // Get the instrument for this order
		    instrumentRequestCh <- InstrumentRequest{in.Instrument, instrumentResultCh}
            instrument := <-instrumentResultCh

            // Send the matching request to the appropriate channel
            if in.OrderType == utils.InputBuy {
                instrument.BuyOrdersCh <- Request{in, matchingResultCh}
            } else {
                instrument.SellOrdersCh <- Request{in, matchingResultCh}
            }

            // Wait till matching request is done and keep track when the order was not fully fulfilled
            matched := <-matchingResultCh
            if !matched {
                unmatchedOrders[in.OrderId] = UnmatchedOrder{in.Instrument, in.OrderType}
            }
        }
	}
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}
