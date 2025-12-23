package main

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Torrent struct {
	Announce     string
	AnnounceList []string
	Info         FileInfo
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

func extractAnnounceURLs(announceListRaw []any) []string {
	var announceList []string

	for _, trackerList := range announceListRaw {
		trackerListSlice, ok := trackerList.([]any)
		if !ok {
			continue
		}

		for _, tracker := range trackerListSlice {
			if str, ok := tracker.(string); ok {
				announceList = append(announceList, str)
			}
		}
	}

	return announceList
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
	if announceList, ok := tRes["announce-list"].([]any); ok {
		torrent.AnnounceList = extractAnnounceURLs(announceList)
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

func (t *Torrent) InfoHash() [20]byte {
	b := BEncode(t.Info)
	hash := sha1.Sum(b)

	return hash
}

func (t *Torrent) buildTrackerURL(
	announceUrl string,
	peerID [20]byte,
	port int,
) (string, error) {
	baseURL, err := url.Parse(announceUrl)
	if err != nil {
		return "", fmt.Errorf("could not parse announce url: %v", err)
	}
	infoHash := t.InfoHash()
	params := url.Values{}
	params.Set("info_hash", string(infoHash[:]))
	params.Set("peer_id", string(peerID[:]))
	params.Set("port", strconv.Itoa(port))
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("left", strconv.FormatInt(t.Info.Length, 10))
	// params.Set("compact", "1")

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

func generatePeerID(prefix string) [20]byte {
	peerID := [20]byte{}

	prefix = fmt.Sprintf("-%s-", prefix)

	copy(peerID[:], prefix)

	rand.Read(peerID[len(prefix):])

	return peerID
}

func (t *Torrent) announceTracker(announceUrl string, peerID [20]byte, port int) ([]Peer, error) {
	trackerURL, err := t.buildTrackerURL(announceUrl, peerID, port)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(trackerURL)
	if err != nil {
		return nil, fmt.Errorf("tracker request failed: %w", err)
	}
	defer resp.Body.Close()

	trackerResp, err := BDecode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tracker response")
	}

	respMap, ok := trackerResp.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid tracker response format")
	}

	if failureReason, ok := respMap["failure reason"].(string); ok {
		return nil, fmt.Errorf("tracker error: %s", failureReason)
	}

	return parsePeers(respMap, announceUrl)
}

func (t *Torrent) getPeers(peerID [20]byte, port int) ([]Peer, error) {
	var allPeers []Peer

	if t.Announce != "" {
		peers, err := t.announceTracker(t.Announce, peerID, port)
		if err == nil && len(peers) > 0 {
			allPeers = append(allPeers, peers...)
		}
	}

	for _, trackerURL := range t.AnnounceList {
		peers, err := t.announceTracker(trackerURL, peerID, port)
		if err == nil && len(peers) > 0 {
			allPeers = append(allPeers, peers...)
		}

	}

	return allPeers, nil
}

func (t *Torrent) getPieceSize(pieceIndex int) int64 {
	lastPieceIndex := len(t.Info.Pieces) - 1
	if pieceIndex == lastPieceIndex {
		if lastSize := t.Info.Length % t.Info.PieceLength; lastSize != 0 {
			return lastSize
		}
	}

	return t.Info.PieceLength
}
