package commands

import (
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bittorrent"
	"log"
	"os"
	"strconv"
)

func DownloadPiece() error {
	if os.Args[2] != "-o" {
		fmt.Println("Output argument not specified")
		os.Exit(1)
	}

	if len(os.Args) != 6 {
		fmt.Println("not enough args")
		os.Exit(1)
	}

	meta, err := bittorrent.ParseMetaInfoFile(os.Args[4])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	trackerInfo, err := bittorrent.FetchTrackerInfo(meta)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	address := trackerInfo.Peers[0].IP.String() + ":" + strconv.FormatInt(int64(trackerInfo.Peers[0].Port), 10)

	conn, err := bittorrent.Dial(meta, address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	payload, err := conn.Receive()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if payload.MessageId == bittorrent.BitFieldMessageId {
		log.Println("BitField received")
	} else {
		log.Printf("Unexpected message id: %v\n", payload.MessageId)
	}

	err = conn.Send(bittorrent.InterestedMessageId, []byte{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Println("Interested message sent")

	payload, err = conn.Receive()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if payload.MessageId == bittorrent.UnchokeMessageId {
		log.Println("Unchoke received")
	} else {
		log.Printf("Unexpected message id: %v\n", payload.MessageId)
	}

	pieceIndex, err := strconv.Atoi(os.Args[5])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data, err := conn.DownloadPiece(meta, pieceIndex)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Write piece to file
	file, err := os.Create(os.Args[3])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()
	file.Write(data)

	fmt.Printf("Piece %d downloaded to %s\n", pieceIndex, os.Args[3])

	return nil
}
