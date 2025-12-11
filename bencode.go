package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
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
		br.UnreadByte()
		return decodeString(br)
	default:
		return nil, fmt.Errorf("invalid bencode type: %c", ch)
	}
}

func decodeInt(br *bufio.Reader) (int64, error) {
	var intBuffer []byte
	for {
		ch, err := br.ReadByte()
		if err != nil {
			return 0, err
		}

		if ch == 'e' {
			i, err := strconv.ParseInt(string(intBuffer), 10, 64)
			if err != nil {
				panic(err)
			}
			return i, nil
		}

		intBuffer = append(intBuffer, ch)
	}
}

func decodeString(br *bufio.Reader) (string, error) {
	var lengthBuffer []byte
	for {
		ch, err := br.ReadByte()
		if err != nil {
			return "", err
		}

		if ch == ':' {
			break
		}

		lengthBuffer = append(lengthBuffer, ch)
	}

	length, err := strconv.ParseInt(string(lengthBuffer), 10, 64)
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(br, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func decodeDict(br *bufio.Reader) (map[string]interface{}, error) {
	return nil, nil
}
func decodeList(br *bufio.Reader) ([]interface{}, error) {
	return nil, nil
}
