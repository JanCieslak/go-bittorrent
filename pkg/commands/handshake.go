package commands

import (
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bittorrent"
	"os"
)

func Handshake() error {
	if len(os.Args) < 3 {
		fmt.Println("No torrent file to parse specified")
		os.Exit(1)
	}

	if len(os.Args) != 4 {
		fmt.Println("no peer specified")
		os.Exit(1)
	}

	meta, err := bittorrent.ParseMetaInfoFile(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := bittorrent.Dial(meta, os.Args[3])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Peer ID:", conn.PeerID)

	return nil
}
