package main

import (
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
