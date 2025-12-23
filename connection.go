package main

import (
	"encoding/binary"
	"net"
	"time"
)

type Connection struct {
	Conn     net.Conn
	Peer     Peer
	Unchoked bool
	Bitfield Bitfield
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewConnection(peer Peer, infoHash, peerID [20]byte) (*Connection, error) {
	conn, err := net.DialTimeout("tcp", peer.Addr(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = performHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bf, err := receiveBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Connection{
		Conn:     conn,
		Peer:     peer,
		Unchoked: false,
		Bitfield: bf,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}

func (c *Connection) SendInterested() error {
	msg := Message{ID: MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Connection) SendRequest(index, begin, length uint32) error {
	msg := Message{ID: MsgRequest}

	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], index)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	binary.BigEndian.PutUint32(payload[8:12], length)

	msg.Payload = payload
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
