package mybittorrent

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

type TrackerInfo struct {
	Interval int
	Peers    []Peer
}

type Peer struct {
	IP   net.IP
	Port int
}

var trackerInfoClient = new(http.Client)

func FetchTrackerInfo(meta *MetaInfoFile) (*TrackerInfo, error) {
	req, err := http.NewRequest(http.MethodGet, meta.Announce, nil)
	if err != nil {
		return nil, err
	}

	infoHash := meta.HashInfo()
	params := req.URL.Query()
	log.Println("info:", url.QueryEscape(string(infoHash[:])))
	params.Add("info_hash", url.QueryEscape(string(infoHash[:])))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.FormatInt(int64(meta.Info.Length), 10))
	params.Add("compact", "1")
	req.URL.RawQuery = params.Encode()

	resp, err := trackerInfoClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	decoded, err := NewDecoder(string(body)).Decode()
	if err != nil {
		return nil, err
	}

	if trackerResponse, ok := decoded.(map[string]interface{}); ok {
		if failure, ok := trackerResponse["failure reason"]; ok {
			return nil, fmt.Errorf("error: %s", failure)
		}

		peers := trackerResponse["peers"].(string)

		if len(peers)%6 != 0 {
			return nil, fmt.Errorf("peers not a multiple of 6 bytes")
		}

		peerList := make([]Peer, 0)
		buf := bytes.NewBufferString(peers)
		for len(buf.Bytes()) > 0 {
			peer := buf.Next(6)
			port := binary.BigEndian.Uint16(peer[4:])
			peerList = append(peerList, Peer{
				IP:   net.IPv4(peer[0], peer[1], peer[2], peer[3]),
				Port: int(port),
			})
		}

		return &TrackerInfo{
			Interval: trackerResponse["interval"].(int),
			Peers:    peerList,
		}, nil
	}

	return nil, fmt.Errorf("invalid tracker response: %s", decoded)
}
