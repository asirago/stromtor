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

func (p *PeerPool) GetBestPeer(index int) *PeerConnection {
	p.mu.Lock()
	defer p.mu.Unlock()

	var bestPeer *PeerConnection
	for _, peerConn := range p.connections {
		if peerConn.Busy {
			continue
		}

		if !peerConn.Conn.Bitfield.HasPiece(index) {
			continue
		}

		if bestPeer == nil || peerConn.Metrics.PiecesFailed <= bestPeer.Metrics.PiecesFailed {
			bestPeer = peerConn
		}

	}

	if bestPeer != nil {
		bestPeer.Busy = true
	}

	return bestPeer
}

func (p *PeerPool) ReleasePeer(peerConn *PeerConnection, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peerConn.Busy = false

	if success {
		peerConn.Metrics.PiecesCompleted++
	} else {
		peerConn.Metrics.PiecesFailed++
	}

	if peerConn.Metrics.PiecesFailed > 5 {
		log.Printf("🗑️ Dropping poor performer: %s", peerConn.Conn.Peer.Addr())
		p.removePeer(peerConn)
	}
}

func (p *PeerPool) removePeer(peerConn *PeerConnection) {
	peerConn.Conn.Conn.Close()

	for i, pc := range p.connections {
		if pc == peerConn {
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			break
		}
	}
}

func (p *PeerPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, peerConn := range p.connections {
		peerConn.Conn.Conn.Close()
	}

	p.connections = nil
}
