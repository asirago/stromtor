package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"time"
)

const MaxRetries = 3

type DownloadState struct {
	Completed Bitfield
	Retries   []int
}

func NewDownloadState(numPieces int) *DownloadState {
	return &DownloadState{
		Completed: NewBitfield(numPieces),
		Retries:   make([]int, numPieces),
	}
}

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

func (t *Torrent) Download(peer Peer, infoHash, peerID [20]byte) error {
	c, err := NewConnection(peer, infoHash, peerID)
	if err != nil {
		return err
	}
	defer c.Conn.Close()

	state := NewDownloadState(t.NumPieces())

	for pieceIdx, pieceHash := range t.Info.Pieces {
		if state.Completed.HasPiece(pieceIdx) {
			continue
		}

		for state.Retries[pieceIdx] < MaxRetries {
			log.Printf(
				"Downloading piece %d (attempt %d/%d)\n",
				pieceIdx,
				state.Retries[pieceIdx]+1,
				MaxRetries,
			)

			// download piece
			downloadedPiece, err := c.DownloadPiece(uint32(pieceIdx), t.getPieceLength(pieceIdx))
			if err != nil {
				state.Retries[pieceIdx]++
				log.Printf("Download failed for piece %d: %v\n", pieceIdx, err)
				continue
			}

			// verify piece hash
			hash := sha1.Sum(downloadedPiece)
			if !bytes.Equal(hash[:], pieceHash[:]) {
				state.Retries[pieceIdx]++
				log.Printf(
					"Hash mismatch for piece %d (attempt %d/%d)\n",
					pieceIdx,
					state.Retries[pieceIdx],
					MaxRetries,
				)
				continue
			}

			// write piece to file
			err = t.writePieceToFile(pieceIdx, downloadedPiece)
			if err != nil {
				return err
			}

			state.Completed.SetPiece(pieceIdx)
			log.Printf(
				"✅ Piece %d of %d verified and written (%.2f%%)\n",
				pieceIdx,
				t.NumPieces(),
				100*(float64(pieceIdx+1)/float64(len(t.Info.Pieces))),
			)

			// success! break out of retry loop
			break
		}

		if !state.Completed.HasPiece(pieceIdx) {
			return fmt.Errorf("failed to download piece %d after %d attempts", pieceIdx, MaxRetries)
		}
	}

	return nil
}

func (t *Torrent) writePieceToFile(pieceIdx int, piece []byte) error {
	pieceOffset := int64(pieceIdx) * t.Info.PieceLength
	fileName := t.Info.Name

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if pieceIdx == 0 {
		if err := file.Truncate(t.Info.Length); err != nil {
			return fmt.Errorf("failed to set file size: %w", err)
		}
	}

	if _, err = file.WriteAt(piece, pieceOffset); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

type PieceWork struct {
	Index  int
	Hash   [20]byte
	Length int
}

type PieceResult struct {
	Index     int
	Buf       []byte
	StartTime time.Time
	EndTime   time.Time
}

func (t *Torrent) DownloadConcurrent(peers []Peer, infoHash, peerID [20]byte) error {
	pool := NewPeerPool(peers, infoHash, peerID, 50)
	pool.ConnectToPeers(peers)
	defer pool.Close()

	retryQueue := make(chan *PieceWork, t.NumPieces())
	workQueue := make(chan *PieceWork, t.NumPieces())
	for pieceIdx, pieceHash := range t.Info.Pieces {
		workQueue <- &PieceWork{
			Index:  pieceIdx,
			Hash:   pieceHash,
			Length: int(t.getPieceLength(pieceIdx)),
		}
	}

	results := make(chan *PieceResult, t.NumPieces())

	numWorkers := 20
	for i := range numWorkers {
		go func(workerID int) {
			log.Printf("🔧 Worker %d: started\n", workerID)

			for {
				var work *PieceWork
				select {
				case work = <-retryQueue:
					log.Printf(
						"🔄 Worker %d: retrying piece %d (HIGH PRIORITY)\n",
						workerID,
						work.Index,
					)
				default:
					select {
					case work = <-retryQueue:
						log.Printf(
							"🔄 Worker %d: retrying piece %d (HIGH PRIORITY)\n",
							workerID,
							work.Index,
						)
					case work = <-workQueue:
					}
				}

				if work == nil {
					break
				}

				peerConn := pool.GetBestPeer(work.Index)
				if peerConn == nil {
					log.Printf(
						"⚠️ Worker %d: no peer available for piece %d, retrying\n",
						workerID,
						work.Index,
					)

					time.Sleep(15 * time.Second)
					retryQueue <- work
					continue
				}

				startTime := time.Now()
				piece, err := peerConn.Conn.DownloadPiece(uint32(work.Index), int64(work.Length))
				if err != nil {
					log.Printf(
						"❌ Worker %d: failed to download piece %d from %s: %v\n",
						workerID,
						work.Index,
						peerConn.Conn.Peer.Addr(),
						err,
					)
					pool.ReleasePeer(peerConn, false)
					retryQueue <- work
					continue
				}

				hash := sha1.Sum(piece)
				if !bytes.Equal(hash[:], work.Hash[:]) {
					log.Printf("❌ Worker %d: hash mismatch for piece %d\n", workerID, work.Index)
					pool.ReleasePeer(peerConn, false)
					retryQueue <- work
					continue
				}

				pool.ReleasePeer(peerConn, true)
				results <- &PieceResult{
					Index:     work.Index,
					Buf:       piece,
					StartTime: startTime,
					EndTime:   time.Now(),
				}
			}

			log.Printf("⏹  Worker %d: shutting down\n", workerID)
		}(i)
	}

	piecesCompleted := 0
	for piecesCompleted < t.NumPieces() {
		result := <-results
		if result.Buf == nil {
			return fmt.Errorf("failed to download piece %d", result.Index)
		}

		err := t.writePieceToFile(result.Index, result.Buf)
		if err != nil {
			return fmt.Errorf("failed to write piece %d: %w", result.Index, err)
		}

		elapsed := result.EndTime.Sub(result.StartTime).Seconds()
		speedKbps := float64(len(result.Buf)) / 1024 / elapsed

		piecesCompleted++
		percent := 100 * float64(piecesCompleted) / float64(t.NumPieces())
		log.Printf(
			"✅ Piece %d complete (%d/%d - %.2f%%) - %.2f KiB/s\n",
			result.Index,
			piecesCompleted,
			t.NumPieces(),
			percent,
			speedKbps,
		)
	}
	close(workQueue)
	close(retryQueue)
	log.Println("🎉 Download complete")

	return nil
}
