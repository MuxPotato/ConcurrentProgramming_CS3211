#include <memory>
#include <mutex>
#include <vector>

#include "instrument.cpp"
#include "io.hpp"

struct Node {
    Instrument instrument;
    std::shared_ptr<Node> next;
    std::mutex m;

    Node(const char* name) : instrument{name}, next{nullptr} {}
};

struct OrderBook {
    // Initialize linked list with dummy node to avoid having to deal with empty list edge case
    std::shared_ptr<Node> head = std::make_shared<Node>(" ");

    Instrument& getInstrument(const char* name) {
        auto curr = head;
        while (true) {
            // Acquire node mutex before reading it
            std::lock_guard lock{curr->m};
            if (curr->instrument.name == name) {
                return curr->instrument;
            }

            if (curr->next == nullptr) {
                // Instrument not found after inspecting the tail node, create a new node
                curr->next = std::make_shared<Node>(name);
                return curr->next->instrument;
            }
            curr = curr->next;
        }
    }

    bool processOrder(ClientCommand& input) {
        Instrument& instrument = getInstrument(input.instrument);
        if (input.type == input_buy) {
            return instrument.processBuyOrder(input);
        }
        return instrument.processSellOrder(input);
    }

    bool cancelOrder(uint32_t id, std::string& instrument, CommandType type) {
        return getInstrument(instrument.c_str()).cancelOrder(id, type);
    }
};
