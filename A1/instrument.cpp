#include <algorithm>
#include <atomic>
#include <condition_variable>
#include <chrono>
#include <deque>
#include <mutex>

#include "io.hpp"

inline std::chrono::microseconds::rep getCurrentTimestamp() noexcept
{
	return std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now().time_since_epoch()).count();
}

// Class to represent a resting order
struct Order {
	uint32_t id;
	uint32_t price;
	uint32_t count;
	uint32_t matchId = 1;
	std::chrono::microseconds::rep timestamp;

	Order(uint32_t id, uint32_t price, uint32_t count, std::chrono::microseconds::rep time) : id{id}, price{price}, count{count}, timestamp{time} {}
};

// Class to represent a successful match against a resting order (only valid when count > 0)
struct Match {
	uint32_t id;
	uint32_t price;
	uint32_t count;
    uint32_t match;
	std::chrono::microseconds::rep timestamp;

    // Match against the best resting order, removing it if its count reaches 0
    void matchOrder(std::deque<Order>& orders, uint32_t& inputCount) {
        id = orders[0].id;
        price = orders[0].price;
        count = std::min(orders[0].count, inputCount);
        match = orders[0].matchId;

        inputCount -= count;
        if (orders[0].count == count) {
            orders.pop_front();
        } else {
            orders[0].count -= count;
            orders[0].matchId++;
        }
        timestamp = getCurrentTimestamp();
    }
};

struct Instrument {
    std::string name;

    std::deque<Order> buyOrders;
    std::deque<Order> sellOrders;
    std::mutex buyOrdersMutex;
    std::mutex sellOrdersMutex;

    std::mutex accessMutex;
    std::condition_variable buyCond;
    std::condition_variable sellCond;
    int buyActive = 0;
    int sellActive = 0;

    Instrument(const char* name) : name{name} {}

    bool processBuyOrder(ClientCommand& input) {
        // Wait until no sell orders active, then increment active buy order count
        std::unique_lock lock{accessMutex};
        while (sellActive > 0) {
            buyCond.wait(lock);
        }
        buyActive++;
        // Unlock mutex so other buy orders can enter
        lock.unlock();

        // Match against best resting order while active order count > 0
        Match matchData;
        do {
            matchData.count = 0;
            {
                std::lock_guard lock{sellOrdersMutex};
                if (sellOrders.empty() || input.price < sellOrders[0].price) {
                    break;
                }
                matchData.matchOrder(sellOrders, input.count);
            }

            // Print output while not holding mutex to allow concurrent processing
            if (matchData.count > 0) {
                Output::OrderExecuted(matchData.id, input.order_id, matchData.match, matchData.price, matchData.count, matchData.timestamp);
            }
        } while (input.count > 0 && matchData.count > 0);

        // Add to order book if remaining count > 0
        if (input.count > 0) {
            std::chrono::microseconds::rep timestamp;
            {
                std::lock_guard lock{buyOrdersMutex};
                timestamp = getCurrentTimestamp();
                buyOrders.push_back(Order{input.order_id, input.price, input.count, timestamp});
                std::sort(buyOrders.begin(), buyOrders.end(), [](const Order& a, const Order& b) {
                    return (a.price == b.price) ? (a.timestamp < b.timestamp) : (a.price > b.price);
                });
            }

            // Print output while not holding mutex to allow concurrent processing
            Output::OrderAdded(input.order_id, input.instrument, input.price, input.count, false, timestamp);
        }
        
        // Decrement active buy order count and signal to selling threads if no buy orders are still processing
        lock.lock();
        buyActive--;
        if (buyActive == 0) {
            sellCond.notify_all();
        }
        lock.unlock();
        return input.count > 0;
    }
    
    bool processSellOrder(ClientCommand& input) {
        // Wait until no buy orders active, then increment active sell order count
        std::unique_lock lock{accessMutex};
        while (buyActive > 0) {
            sellCond.wait(lock);
        }
        sellActive++;
        // Unlock mutex so other sell orders can enter
        lock.unlock();

        // Match against best resting order while active order count > 0
        Match matchData;
        do {
            matchData.count = 0;
            {
                std::lock_guard lock{buyOrdersMutex};
                if (buyOrders.empty() || input.price > buyOrders[0].price) {
                    break;
                }
                matchData.matchOrder(buyOrders, input.count);
            }

            // Print output while not holding mutex to allow concurrent processing
            if (matchData.count > 0) {
                Output::OrderExecuted(matchData.id, input.order_id, matchData.match, matchData.price, matchData.count, matchData.timestamp);
            }
        } while (input.count > 0 && matchData.count > 0);

        // Add to order book if remaining count > 0
        if (input.count > 0) {
            std::chrono::microseconds::rep timestamp;
            {
                std::lock_guard lock{sellOrdersMutex};
                timestamp = getCurrentTimestamp();
                sellOrders.push_back(Order{input.order_id, input.price, input.count, getCurrentTimestamp()});
                std::sort(sellOrders.begin(), sellOrders.end(), [](const Order& a, const Order& b) {
                    return (a.price == b.price) ? (a.timestamp < b.timestamp) : (a.price < b.price);
                });
            }

            // Print output while not holding mutex to allow concurrent processing
            Output::OrderAdded(input.order_id, input.instrument, input.price, input.count, true, timestamp);
        }
        
        // Decrement active sell order count and signal to buying threads if no sell orders are still processing
        lock.lock();
        sellActive--;
        if (sellActive == 0) {
            buyCond.notify_all();
        }
        lock.unlock();
        return input.count > 0;
    }

    // Returns whether order was found in the order book, and hence, successfully cancelled
    bool cancelOrder(uint32_t id, CommandType type) {
        std::lock_guard lock{type == input_buy ? buyOrdersMutex : sellOrdersMutex};
        std::deque<Order>& orders = (type == input_buy ? buyOrders : sellOrders);

        for (auto it{orders.begin()}; it != orders.end(); it++) {
            if (it->id == id) {
                orders.erase(it);
                return true;
            }
        }
        return false;
    }
};
