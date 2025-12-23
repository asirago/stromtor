package main

import "log"

type PieceProgress struct {
	Index      uint32
	PeerConn   *Connection
	Buf        []byte
	Downloaded int
	Requested  int
}

func NewPieceProgress(index uint32, length int64, conn *Connection) *PieceProgress {
	return &PieceProgress{
		Index:    index,
		Buf:      make([]byte, length),
		PeerConn: conn,
	}
}

func (pp *PieceProgress) readMessage() error {
	msg, err := ReadMessage(pp.PeerConn.Conn)
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case MsgUnchoke:
		pp.PeerConn.Unchoked = true
	case MsgChoke:
		pp.PeerConn.Unchoked = false
	case MsgPiece:
		n, err := ParsePiece(pp.Index, pp.Buf, msg)
		if err != nil {
			return err
		}
		pp.Downloaded += n
	default:
		log.Printf("Received: Message ID %d\n", msg.ID)
	}

	return nil
}
