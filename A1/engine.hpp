// This file contains declarations for the main Engine class. You will
// need to add declarations to this file as you develop your Engine.

#ifndef ENGINE_HPP
#define ENGINE_HPP

#include <chrono>
#include <mutex>
#include <deque>
#include <unordered_map>

#include "io.hpp"
#include "order_book.cpp"

struct Engine
{
public:
	void accept(ClientConnection conn);

private:
	OrderBook orderBook;
	void connection_thread(ClientConnection conn);
};

#endif
