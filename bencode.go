package main

import (
	"bufio"
	"fmt"
	"io"
)

func BDecode(r io.Reader) (interface{}, error) {
	br := bufio.NewReader(r)

	ch, err := br.ReadByte()
	if err != nil {
		return nil, err
	}

	switch {
	case ch == 'i':
		return decodeInt(br)
	case ch == 'd':
		return decodeDict(br)
	case ch == 'l':
		return decodeList(br)
	case ch >= '0' && ch <= '9':
		return decodeString(br)
	default:
		return nil, fmt.Errorf("invalid bencode type: %c", ch)
	}
}

func decodeInt(br *bufio.Reader) (int64, error) {
	return 0, nil
}

func decodeString(br *bufio.Reader) (string, error) {
	return "", nil
}

func decodeDict(br *bufio.Reader) (map[string]interface{}, error) {
	return nil, nil
}
func decodeList(br *bufio.Reader) ([]interface{}, error) {
	return nil, nil
}
