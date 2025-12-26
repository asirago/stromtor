package main

import (
	"sync"
)

type PeerMetrics struct {
	PiecesCompleted int
	PiecesFailed    int
}

type PeerConnection struct {
	Conn    *Connection
	Metrics PeerMetrics
	Busy    bool
}

type PeerPool struct {
	mu          sync.Mutex
	connections []*PeerConnection
	infoHash    [20]byte
	peerID      [20]byte
	maxPeers    int
}

func NewPeerPool(peers []Peer, infoHash, peerID [20]byte, maxPeers int) *PeerPool {
	return &PeerPool{
		connections: make([]*PeerConnection, 0, maxPeers),
		infoHash:    infoHash,
		peerID:      peerID,
		maxPeers:    maxPeers,
	}
}
