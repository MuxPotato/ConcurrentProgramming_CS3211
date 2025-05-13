package utils

// The cgo code below interfaces with the struct in io.h
// There should be no need to modify this file.

/*
#include <stdint.h>
#include "../io.h"

// Capitalized to export.
// Do not use lower caps.
typedef struct {
	enum CommandType Type;
	uint32_t Order_id;
	uint32_t Price;
	uint32_t Count;
	char Instrument[9];
}cInput;
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"
)

type Input struct {
	OrderType  InputType
	OrderId    uint32
	Price      uint32
	Count      uint32
	Instrument string
}

type InputType byte

const (
	InputBuy    InputType = 'B'
	InputSell   InputType = 'S'
	InputCancel InputType = 'C'
)

func ReadInput(conn net.Conn) (in Input, err error) {
	buf := make([]byte, unsafe.Sizeof(C.cInput{}))
	_, err = conn.Read(buf)
	if err != nil {
		return
	}

	var cin C.cInput
	b := bytes.NewBuffer(buf)
	err = binary.Read(b, binary.LittleEndian, &cin)
	if err != nil {
		fmt.Printf("read err: %v\n", err)
		return
	}

	in.OrderType = (InputType)(cin.Type)
	in.OrderId = (uint32)(cin.Order_id)
	in.Price = (uint32)(cin.Price)
	in.Count = (uint32)(cin.Count)

	len := 0
	tmp := make([]byte, 9)
	for i, c := range cin.Instrument {
		tmp[i] = (byte)(c)
		if c != 0 {
			len++
		}
	}

	in.Instrument = string(tmp[:len])
	// in.instrument = *(*string)(unsafe.Pointer(&tmp))

	return
}

func OutputOrderDeleted(in Input, accepted bool, outTime int64) {
	acceptedTxt := "A"
	if !accepted {
		acceptedTxt = "R"
	}
	fmt.Printf("X %v %v %v\n",
		in.OrderId, acceptedTxt, outTime)
}

func OutputOrderAdded(in Input, outTime int64) {
	orderType := "S"
	if in.OrderType == InputBuy {
		orderType = "B"
	}
	fmt.Printf("%v %v %v %v %v %v\n",
		orderType, in.OrderId, in.Instrument, in.Price, in.Count, outTime)
}

func OutputOrderExecuted(restingId, newId, execId, price, count uint32, outTime int64) {
	fmt.Printf("E %v %v %v %v %v %v\n",
		restingId, newId, execId, price, count, outTime)
}
