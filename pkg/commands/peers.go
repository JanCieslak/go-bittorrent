package commands

import (
	"errors"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/bittorrent"
	"os"
	"strconv"
)

func Peers() error {
	if len(os.Args) != 3 {
		return errors.New("no torrent file to parse specified")
	}

	meta, err := bittorrent.ParseMetaInfoFile(os.Args[2])
	if err != nil {
		return err
	}

	trackerInfo, err := bittorrent.FetchTrackerInfo(meta)
	if err != nil {
		return err
	}

	for _, peer := range trackerInfo.Peers {
		fmt.Println(peer.IP.String() + ":" + strconv.FormatInt(int64(peer.Port), 10))
	}

	return nil
}
