package main

import (
	"bytes"
	"log"
	"os"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing <filename>\n")
	}

	filename := os.Args[1]

	// read torrent file
	data, err := os.ReadFile(filename)
	check(err)

	// parse torrent file
	torrent, err := parseTorrent(bytes.NewReader(data))
	check(err)

	// generate peer id
	peerID := generatePeerID("ST001")

	// get peers from tracker
	var listOfPeers []Peer
	listOfPeers, err = torrent.getPeers(peerID, 6881)
	check(err)

	// start concurrent download
	err = torrent.DownloadConcurrent(listOfPeers, torrent.InfoHash(), peerID)
	check(err)

	log.Println("Torrent Client Just Ended ")
}
