package commands

import (
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bittorrent"
	"log"
	"os"
	"strconv"
)

func Download() error {
	if os.Args[2] != "-o" {
		fmt.Println("Output argument not specified")
		os.Exit(1)
	}

	if len(os.Args) != 5 {
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

	data, err := conn.Download(meta)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = os.WriteFile(os.Args[3], data, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded %s to %s\n", os.Args[4], os.Args[3])

	return nil
}
