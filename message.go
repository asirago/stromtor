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
