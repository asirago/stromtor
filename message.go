package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

type MessageType uint8

const (
	MsgChoke MessageType = iota
	MsgUnchoke
	MsgInterested
	MsgNotInterested
	MsgHave
	MsgBitfield
	MsgRequest
	MsgPiece
	MsgCancel
)

type Bitfield []byte

func NewBitfield(numPieces int64) *Bitfield {
	bitfieldSize := (numPieces + 7) / 8

	bitfield := Bitfield(make([]byte, bitfieldSize))

	return &bitfield
}

func (bf *Bitfield) SetPiece(index int64) {
	byteIndex := index / 8
	byteOffset := index % 8

	(*bf)[byteIndex] |= 1 << (7 - byteOffset)
}

func (bf *Bitfield) HasPiece(index int64) bool {
	byteIndex := index / 8
	byteOffset := index % 8

	return (*bf)[byteIndex]&(1<<(7-byteOffset)) != 0
}

type Message struct {
	ID      MessageType
	Payload []byte
}

func WriteMessage(conn net.Conn, msg *Message) error {
	var buf bytes.Buffer

	length := uint32(1 + len(msg.Payload))
	binary.Write(&buf, binary.BigEndian, length)
	buf.WriteByte(byte(msg.ID))
	buf.Write(msg.Payload)

	_, err := conn.Write(buf.Bytes())
	return err
}

func ReadMessage(conn net.Conn) (*Message, error) {

	lengthMsg := make([]byte, 4)
	_, err := conn.Read(lengthMsg)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthMsg)
	if length == 0 {
		return &Message{ID: 255}, nil
	}

	msgBuf := make([]byte, length)
	_, err = io.ReadFull(conn, msgBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      MessageType(msgBuf[0]),
		Payload: msgBuf[1:],
	}, nil

}
