package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

type Peer struct {
	IP         net.IP
	Port       uint16
	TrackerURl string
	Discovered time.Time
}

func parsePeers(peersResp map[string]any, trackerURL string) ([]Peer, error) {
	if peers, ok := peersResp["peers"].(string); ok {
		return parseCompactPeers(peers, trackerURL), nil
	}

	if peers, ok := peersResp["peers"].([]any); ok {
		return parseNonCompactPeers(peers, trackerURL), nil
	}

	return nil, fmt.Errorf("no peers in tracker response")
}

func parseNonCompactPeers(peers []any, trackerURL string) []Peer {
	var listPeers []Peer

	for _, peer := range peers {
		peerData := peer.(map[string]any)

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
			IP:         parsedIP,
			Port:       uint16(port),
			TrackerURl: trackerURL,
			Discovered: time.Now(),
		})
	}

	return listPeers
}

func parseCompactPeers(peers, trackerURL string) []Peer {
	const peerSize = 6
	numPeers := len(peers) / peerSize
	listPeers := make([]Peer, numPeers)

	for i := range numPeers {
		offset := i * peerSize
		peerBytes := peers[offset : offset+peerSize]
		listPeers[i] = Peer{
			IP:         net.IPv4(peerBytes[0], peerBytes[1], peerBytes[2], peerBytes[3]),
			Port:       binary.BigEndian.Uint16([]byte(peerBytes[4:6])),
			TrackerURl: trackerURL,
			Discovered: time.Now(),
		}
	}
	return listPeers
}

func (p Peer) Addr() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
