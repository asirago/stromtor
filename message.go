package main

type MessageType uint8

const (
	MsgChoke MessageType = iota
	MsgUnchoke
	MsgInterested
	MsgNotInterested
	MsgHave
	MsgBitfield
	MsgRequest
	MsgPiece
	MsgCancel
)

type Bitfield []byte

func NewBitfield(numPieces int64) *Bitfield {
	bitfieldSize := (numPieces + 7) / 8

	bitfield := Bitfield(make([]byte, bitfieldSize))

	return &bitfield
}

func (bf *Bitfield) SetPiece(index int64) {
	byteIndex := index / 8
	byteOffset := index % 8

	(*bf)[byteIndex] |= 1 << (7 - byteOffset)
}

func (bf *Bitfield) HasPiece(index int64) bool {
	byteIndex := index / 8
	byteOffset := index % 8

	return (*bf)[byteIndex]&(1<<(7-byteOffset)) != 0
}
