package main

import (
	"encoding/binary"
	"net"
)

type Peer struct {
	IP     net.IP
	Port   uint16
	PeerID [20]byte
}

func parsePeers(peers any) []Peer {
	peerList := peers.([]any)
	var listPeers []Peer

	for _, peer := range peerList {
		peerData := peer.(map[string]any)

		var peerID [20]byte
		copy(peerID[:], peerData["peer id"].(string))

		listPeers = append(listPeers, Peer{
			IP:     net.ParseIP(peerData["ip"].(string)),
			Port:   uint16(peerData["port"].(int64)),
			PeerID: peerID,
		})
	}

	return listPeers
}

func parseCompactPeers(peers string) []Peer {
	const peerSize = 6
	numPeers := len(peers) / peerSize
	listPeers := make([]Peer, numPeers)

	for i := range numPeers {
		offset := i * peerSize
		peerBytes := peers[offset : offset+peerSize]
		peer := Peer{
			IP:   net.IPv4(peerBytes[0], peerBytes[1], peerBytes[2], peerBytes[3]),
			Port: binary.BigEndian.Uint16([]byte(peerBytes[4:6])),
		}
		listPeers = append(listPeers, peer)
	}
	return listPeers
}
