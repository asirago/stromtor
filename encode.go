package main

import (
	"bytes"
	"strconv"
)

func BEncode(v any) []byte {
	var buf bytes.Buffer
	bencode(&buf, v)
	return buf.Bytes()
}

func bencode(buf *bytes.Buffer, v any) {
	switch val := v.(type) {
	case int64:
		buf.WriteByte('i')
		buf.WriteString(strconv.FormatInt(val, 10))
		buf.WriteByte('e')
	case string:
		buf.WriteString(strconv.Itoa(len(val)))
		buf.WriteByte(':')
		buf.WriteString(val)
	case FileInfo:
		encodeFileInfo(buf, val)
	}
}

func encodeFileInfo(buf *bytes.Buffer, info FileInfo) {
	buf.WriteByte('d')

	if len(info.Files) > 0 {
		bencode(buf, "files")
		buf.WriteByte('l')
		for _, f := range info.Files {
			buf.WriteByte('d')
			bencode(buf, "length")
			bencode(buf, f.Length)

			bencode(buf, "path")
			buf.WriteByte('l')
			for _, p := range f.Path {
				bencode(buf, p)
			}
			buf.WriteByte('e')

			buf.WriteByte('e')
		}
		buf.WriteByte('e')
	}

	if info.Length > 0 {
		bencode(buf, "length")
		bencode(buf, info.Length)
	}

	bencode(buf, "name")
	bencode(buf, info.Name)

	bencode(buf, "piece length")
	bencode(buf, info.PieceLength)

	bencode(buf, "pieces")
	var piecesBuf bytes.Buffer

	for _, piece := range info.Pieces {
		piecesBuf.WriteString(string(piece[:]))
	}
	bencode(buf, piecesBuf.String())

	if info.Private != 0 {
		bencode(buf, "private")
		bencode(buf, info.Private)
	}

	if info.Source != "" {
		bencode(buf, "source")
		bencode(buf, info.Source)
	}

	buf.WriteByte('e')
}
