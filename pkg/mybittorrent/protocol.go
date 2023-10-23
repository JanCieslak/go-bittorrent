package mybittorrent

import (
	"bytes"
	"encoding/hex"
	"log"
	"net"
)

const BitTorrentProtocolHeader = "BitTorrent protocol"

func HandshakePeer(meta *MetaInfoFile, address string) (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return "", err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return "", err
	}

	peerId, err := randomPeerId()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.WriteByte(19)
	buf.WriteString(BitTorrentProtocolHeader)
	buf.Write(make([]byte, 8))
	buf.Write(meta.HashInfo())
	buf.Write(peerId)
	log.Println("Sending:", buf.String())
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	peerBuf := make([]byte, 68)
	_, err = conn.Read(peerBuf)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(peerBuf[48:]), nil
}
