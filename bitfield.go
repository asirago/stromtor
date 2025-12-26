package main

import (
	"fmt"
	"net"
	"time"
)

type Bitfield []byte

func NewBitfield(numPieces int) Bitfield {
	bitfieldSize := (numPieces + 7) / 8

	bitfield := Bitfield(make([]byte, bitfieldSize))

	return bitfield
}

func (bf *Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	byteOffset := index % 8

	(*bf)[byteIndex] |= 1 << (7 - byteOffset)
}

func (bf *Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	byteOffset := index % 8

	return (*bf)[byteIndex]&(1<<(7-byteOffset)) != 0
}

func receiveBitfield(conn net.Conn) (Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})
	msg, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}

	if msg.ID != MsgBitfield {
		return nil, fmt.Errorf("expected bitfield, but got ID %d", msg.ID)
	}
	if msg == nil {
		return nil, fmt.Errorf("expected bitfield, but got %v", msg)
	}

	return msg.Payload, nil
}
