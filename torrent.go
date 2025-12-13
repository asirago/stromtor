package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

type Torrent struct {
	Announce string
	Info     FileInfo
}

type FileInfo struct {
	Files       []File
	Length      int64
	Name        string
	PieceLength int64
	Pieces      [][20]byte
	Private     int64
	Source      string
}

type File struct {
	Length int64
	Path   []string
}

func extractPieceHashes(pieces string) [][20]byte {
	numHashes := len(pieces) / 20

	result := make([][20]byte, numHashes)
	for i := range numHashes {
		copy(result[i][:], pieces[i*20:(i+1)*20])
	}

	return result
}

func parseTorrent(r io.Reader) (*Torrent, error) {
	res, err := BDecode(r)
	if err != nil {
		return nil, err
	}

	tRes := res.(map[string]any)
	tInfo := tRes["info"].(map[string]any)

	torrent := &Torrent{
		Announce: tRes["announce"].(string),
		Info: FileInfo{
			Name:        tInfo["name"].(string),
			PieceLength: tInfo["piece length"].(int64),
			Pieces:      extractPieceHashes(tInfo["pieces"].(string)),
		},
	}

	if length, ok := tInfo["length"].(int64); ok {
		torrent.Info.Length = length
	} else if files, ok := tInfo["files"].([]any); ok {
		torrent.Info.Files = make([]File, len(files))
		for i, f := range files {
			fileDict := f.(map[string]any)

			pathList := fileDict["path"].([]any)
			pathStrs := make([]string, len(pathList))
			for j, pathStr := range pathList {
				pathStrs[j] = pathStr.(string)
			}

			torrent.Info.Files[i] = File{
				Length: fileDict["length"].(int64),
				Path:   pathStrs,
			}
		}
	}

	if private, ok := tInfo["private"]; ok {
		torrent.Info.Private = private.(int64)
	}

	if source, ok := tInfo["source"]; ok {
		torrent.Info.Source = source.(string)
	}

	return torrent, nil
}

func (t *Torrent) InfoHash() []byte {
	b := BEncode(t.Info)
	hash := sha1.Sum(b)

	return hash[:]
}

func (t *Torrent) buildTrackerURL(
	peerID string,
	port int,
	uploaded, downloaded, left int64,
) (string, error) {
	baseURL, err := url.Parse(t.Announce)
	if err != nil {
		return "", fmt.Errorf("could not parse announce url: %v", err)
	}
	params := url.Values{}
	params.Set("info_hash", string(t.InfoHash()))
	params.Set("peer_id", peerID)
	params.Set("port", strconv.Itoa(port))
	params.Set("uploaded", strconv.FormatInt(uploaded, 10))
	params.Set("downloaded", strconv.FormatInt(downloaded, 10))
	params.Set("left", strconv.FormatInt(left, 10))

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}
