package main

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
