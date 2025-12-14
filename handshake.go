package main

import "bytes"

type Handshake struct {
	Prefix   string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Prefix:   "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	msg := &bytes.Buffer{}
	msg.WriteByte(19)
	msg.Write([]byte(h.Prefix))
	msg.Write(make([]byte, 8))
	msg.Write(h.InfoHash[:])
	msg.Write(h.PeerID[:])

	return msg.Bytes()
}
