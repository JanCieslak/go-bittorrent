package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/mybittorrent/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/mybittorrent/torrent"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) <= 2 {
		fmt.Println("Command not specified")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "decode":
		if len(os.Args) != 3 {
			fmt.Println("No value to decode specified")
			os.Exit(1)
		}

		decoded, err := bencode.NewDecoder(os.Args[2]).Decode()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	case "info":
		if len(os.Args) != 3 {
			fmt.Println("No torrent file to parse specified")
			os.Exit(1)
		}

		meta, err := torrent.ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Tracker URL:", meta.Announce)
		fmt.Println("Length:", meta.Info.Length)
		fmt.Println("Info Hash:", hex.EncodeToString(meta.HashInfo()))
		fmt.Println("Piece Length:", meta.Info.PieceLength)
		fmt.Println("Piece Hashes:")
		for _, p := range meta.Info.Pieces {
			fmt.Println(p)
		}

	case "peers":
		if len(os.Args) != 3 {
			fmt.Println("No torrent file to parse specified")
			os.Exit(1)
		}

		meta, err := torrent.ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		trackerInfo, err := torrent.FetchTrackerInfo(meta)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, peer := range trackerInfo.Peers {
			fmt.Println(peer.IP.String() + ":" + strconv.FormatInt(int64(peer.Port), 10))
		}
	case "handshake":
		if len(os.Args) < 3 {
			fmt.Println("No torrent file to parse specified")
			os.Exit(1)
		}

		if len(os.Args) != 4 {
			fmt.Println("no peer specified")
			os.Exit(1)
		}

		meta, err := torrent.ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		conn, err := torrent.EstablishPeerConnection(meta, os.Args[3])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Peer ID:", conn.PeerID)
	case "download_piece":
		if os.Args[2] != "-o" {
			fmt.Println("Output argument not specified")
			os.Exit(1)
		}

		if len(os.Args) != 6 {
			fmt.Println("not enough args")
			os.Exit(1)
		}

		meta, err := torrent.ParseMetaInfoFile(os.Args[4])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		trackerInfo, err := torrent.FetchTrackerInfo(meta)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		address := trackerInfo.Peers[0].IP.String() + ":" + strconv.FormatInt(int64(trackerInfo.Peers[0].Port), 10)

		conn, err := torrent.EstablishPeerConnection(meta, address)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		payload, err := conn.Receive()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if payload.MessageId == torrent.BitFieldMessageId {
			log.Println("BitField received")
		} else {
			log.Printf("Unexpected message id: %v\n", payload.MessageId)
		}

		err = conn.Send(torrent.InterestedMessageId, []byte{})
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

		if payload.MessageId == torrent.UnchokeMessageId {
			log.Println("Unchoke received")
		} else {
			log.Printf("Unexpected message id: %v\n", payload.MessageId)
		}

		pieceIndex, err := strconv.Atoi(os.Args[5])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		data, err := conn.Download(meta, pieceIndex)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Write piece to file
		file, err := os.Create(os.Args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		file.Write(data)

		fmt.Printf("Piece %d downloaded to %s\n", pieceIndex, os.Args[3])
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
