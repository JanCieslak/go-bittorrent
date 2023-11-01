package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bittorrent"
	"os"
)

func Info() error {
	if len(os.Args) != 3 {
		return errors.New("no torrent file to parse specified")
	}

	meta, err := bittorrent.ParseMetaInfoFile(os.Args[2])
	if err != nil {
		return err
	}

	fmt.Println("Tracker URL:", meta.Announce)
	fmt.Println("Length:", meta.Info.Length)
	fmt.Println("Info Hash:", hex.EncodeToString(meta.HashInfo()))
	fmt.Println("Piece Length:", meta.Info.PieceLength)
	fmt.Println("Piece Hashes:")
	for _, p := range meta.Info.Pieces {
		fmt.Println(p)
	}

	return nil
}
