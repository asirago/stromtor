package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

type Message struct {
	ID      MessageType
	Payload []byte
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	var buf bytes.Buffer
	length := uint32(len(m.Payload) + 1)
	binary.Write(&buf, binary.BigEndian, length)
	buf.WriteByte(byte(m.ID))
	buf.Write(m.Payload)

	return buf.Bytes()
}

func ReadMessage(conn net.Conn) (*Message, error) {
	lengthMsg := make([]byte, 4)
	_, err := io.ReadFull(conn, lengthMsg)
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

func ParsePiece(index uint32, buf []byte, msg *Message) (int, error) {
	parsedIndex := binary.BigEndian.Uint32(msg.Payload[0:4])
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	block := msg.Payload[8:]
	if begin+len(block) > len(buf) {
		return 0, fmt.Errorf(
			"block too long %d for offset %d with length %d",
			len(block),
			begin,
			len(buf),
		)
	}

	copy(buf[begin:], block)
	return len(block), nil
}
