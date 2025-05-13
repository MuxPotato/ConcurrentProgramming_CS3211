#include <algorithm>
#include <iostream>
#include <thread>
#include <unordered_map>
#include "io.hpp"
#include "engine.hpp"

void Engine::accept(ClientConnection connection)
{
	auto thread = std::thread(&Engine::connection_thread, this, std::move(connection));
	thread.detach();
}

void Engine::connection_thread(ClientConnection connection)
{
	// Keep track of what orders this client sent in, since you cannot cancel others' orders
	std::unordered_map<uint32_t, std::pair<std::string, CommandType>> orders;

	while(true)
	{
		ClientCommand input {};
		switch(connection.readInput(input))
		{
			case ReadResult::Error: SyncCerr {} << "Error reading input" << std::endl;
			case ReadResult::EndOfFile: return;
			case ReadResult::Success: break;
		}

		// Functions for printing output actions in the prescribed format are
		// provided in the Output class:
		switch(input.type)
		{
			case input_cancel: {
				SyncCerr {} << "Got cancel: ID: " << input.order_id << std::endl;

				uint32_t id = input.order_id;
				bool cancelled = orders.count(id) > 0 && orderBook.cancelOrder(id, orders[id].first, orders[id].second);
				Output::OrderDeleted(input.order_id, cancelled, getCurrentTimestamp());
				if (cancelled) {
					orders.erase(id);
				}

				break;
			}

			default: {
				SyncCerr {} << "Got order: " << static_cast<char>(input.type) << " " << input.instrument << " x " << input.count << " @ " << input.price << " ID: " << input.order_id << std::endl;

				if (orderBook.processOrder(input)) {
					orders[input.order_id] = {input.instrument, input.type};
				}
				
				break;
			}
		}
	}
}
