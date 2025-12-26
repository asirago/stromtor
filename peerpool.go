package main

import (
	"log"
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

func (p *PeerPool) ConnectToPeers(peers []Peer) {
	var wg sync.WaitGroup
	connectionsChan := make(chan *PeerConnection, len(peers))
	for _, peer := range peers {
		if len(p.connections) >= p.maxPeers {
			break
		}

		wg.Add(1)
		go func(peer Peer) {
			defer wg.Done()
			conn, err := NewConnection(peer, p.infoHash, p.peerID)
			if err != nil {
				log.Printf("Failed to connect to %s: %v\n", peer.Addr(), err)
				return
			}

			peerConn := &PeerConnection{
				Conn: conn,
				Busy: false,
			}

			connectionsChan <- peerConn
		}(peer)
	}

	wg.Wait()
	close(connectionsChan)

	for peerConn := range connectionsChan {
		p.mu.Lock()
		if len(p.connections) < p.maxPeers {
			p.connections = append(p.connections, peerConn)
			log.Printf(
				"✅ Connected to %s (%d/%d)\n",
				peerConn.Conn.Peer.Addr(),
				len(p.connections),
				p.maxPeers,
			)
		}
		p.mu.Unlock()
	}

	log.Printf("🔗 Peer pool initialized with %d connection\n", len(p.connections))
}
