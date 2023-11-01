package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bittorrent/bencoding"
	"os"
)

func Decode() error {
	if len(os.Args) != 3 {
		return errors.New("no value to decode specified")
	}

	decoded, err := bencoding.DecodeString(os.Args[2])
	if err != nil {
		return err
	}

	jsonOutput, err := json.Marshal(decoded)
	if err != nil {
		return err
	}

	fmt.Println(string(jsonOutput))

	return nil
}
