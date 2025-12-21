package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	IP     net.IP
	Port   uint16
	PeerID [20]byte
}

func parsePeers(peersResp map[string]any) ([]Peer, error) {
	if peers, ok := peersResp["peers"].(string); ok {
		return parseCompactPeers(peers), nil
	}

	if peers, ok := peersResp["peers"].([]any); ok {
		return parseNonCompactPeers(peers), nil
	}

	return nil, fmt.Errorf("no peers in tracker response")
}

func parseNonCompactPeers(peers []any) []Peer {
	var listPeers []Peer

	for _, peer := range peers {
		peerData := peer.(map[string]any)

		var peerID [20]byte
		if id, ok := peerData["peer id"].(string); ok {
			copy(peerID[:], id)
		}
		ip, ok := peerData["ip"].(string)
		if !ok {
			continue
		}
		parsedIP := net.ParseIP(ip)

		port, ok := peerData["port"].(int64)
		if !ok || port < 0 || port > 65535 {
			continue
		}

		listPeers = append(listPeers, Peer{
			IP:     parsedIP,
			Port:   uint16(port),
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
		listPeers[i] = Peer{
			IP:   net.IPv4(peerBytes[0], peerBytes[1], peerBytes[2], peerBytes[3]),
			Port: binary.BigEndian.Uint16([]byte(peerBytes[4:6])),
		}
	}
	return listPeers
}

func (p Peer) Addr() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
