// This file contains definitions used by the provided I/O code.
// There should be no need to modify this file.

#ifndef IO_H
#define IO_H

#include <stdint.h>

enum CommandType { input_buy = 'B', input_sell = 'S', input_cancel = 'C' };

struct ClientCommand {
  enum CommandType type;
  uint32_t order_id;
  uint32_t price;
  uint32_t count;
  char instrument[9];
};

#endif
