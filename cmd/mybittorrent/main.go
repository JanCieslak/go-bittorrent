package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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

		decoded, err := NewDecoder(os.Args[2]).Decode()
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

		meta, err := ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		infoHash := meta.HashInfo()
		fmt.Println("Tracker URL:", meta.Announce)
		fmt.Println("Length:", meta.Info.Length)
		fmt.Println("Info Hash:", hex.EncodeToString(infoHash[:]))
		fmt.Println("Piece Length:", meta.Info.PieceLength)
		fmt.Println("Piece Hashes:")
		for _, p := range meta.Info.Pieces {
			fmt.Println(p)
		}

		trackerInfo, err := FetchTrackerInfo(meta)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, peer := range trackerInfo.Peers {
			fmt.Println(peer.IP.String() + ":" + strconv.FormatInt(int64(peer.Port), 10))
		}
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
