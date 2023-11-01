package main

import (
	"errors"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/commands"
	"log"
	"os"
)

func exit(message string) {
	log.Println(message)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		exit("No command specified")
	}

	command := os.Args[1]
	if err := runCommand(command); err != nil {
		exit(err.Error())
	}
}

func runCommand(command string) error {
	switch command {
	case "decode":
		return commands.Decode()
	case "info":
		return commands.Info()
	case "peers":
		return commands.Peers()
	case "handshake":
		return commands.Handshake()
	case "download_piece":
		return commands.DownloadPiece()
	case "download":
		return commands.Download()
	default:
		return errors.New("unknown command: " + command)
	}
}
