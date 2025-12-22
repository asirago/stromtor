package main

import (
	"bytes"
	"fmt"
	"io"
)

type Handshake struct {
	Length   uint8
	Prefix   string
	Reserved [8]byte
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Length:   byte(19),
		Prefix:   "BitTorrent protocol",
		Reserved: [8]byte{},
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
	buf := make([]byte, 68)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	var reserved [8]byte
	copy(reserved[:], buf[20:28])

	var infoHash [20]byte
	copy(infoHash[:], buf[28:48])

	var peerID [20]byte
	copy(peerID[:], buf[48:68])

	return &Handshake{
		Length:   buf[0],
		Prefix:   string(buf[1:20]),
		Reserved: reserved,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}

func (h *Handshake) PrintHandshake() {
	fmt.Printf("Handshake received:\n")
	fmt.Printf("  Protocol Length: \\x%02x\n", h.Length)
	fmt.Printf("  Protocol: %s\n", h.Prefix)
	fmt.Printf("  Reserved: %x\n", h.Reserved)
	fmt.Printf("  Info Hash: %x\n", h.InfoHash)
	fmt.Printf("  Peer ID: %s\n", h.PeerID)
}
func (h *Handshake) Serialize() []byte {
	msg := &bytes.Buffer{}
	msg.WriteByte(h.Length)
	msg.Write([]byte(h.Prefix))
	msg.Write(h.Reserved[:])
	msg.Write(h.InfoHash[:])
	msg.Write(h.PeerID[:])

	return msg.Bytes()
}
