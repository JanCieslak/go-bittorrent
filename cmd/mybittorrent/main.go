package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/mybittorrent"
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

		decoded, err := mybittorrent.NewDecoder(os.Args[2]).Decode()
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

		meta, err := mybittorrent.ParseMetaInfoFile(os.Args[2])
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

		meta, err := mybittorrent.ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		trackerInfo, err := mybittorrent.FetchTrackerInfo(meta)
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

		meta, err := mybittorrent.ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		peerId, err := mybittorrent.HandshakePeer(meta, os.Args[3])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Peer ID:", peerId)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
