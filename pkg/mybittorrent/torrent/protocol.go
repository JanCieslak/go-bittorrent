package torrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/pkg/mybittorrent/internal"
	"io"
	"log"
	"net"
)

const BitTorrentProtocolHeader = "BitTorrent protocol"

type MessageId byte

const (
	UnchokeMessageId    MessageId = 1
	InterestedMessageId MessageId = 2
	BitFieldMessageId   MessageId = 5
	RequestMessageId    MessageId = 6
	PieceMessageId      MessageId = 7

	BlockSize = 16 * 1024
)

type Payload struct {
	Len       uint32
	MessageId MessageId
	Data      []byte
}

type PeerConn struct {
	PeerID string
	conn   *net.TCPConn
}

func EstablishPeerConnection(meta *MetaInfoFile, address string) (*PeerConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	peerId, err := internal.RandomPeerId()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	peerBuf := make([]byte, 68)
	_, err = conn.Read(peerBuf)
	if err != nil {
		return nil, fmt.Errorf("conn read: %w", err)
	}

	log.Println("Connection established successfully!")

	return &PeerConn{
		PeerID: hex.EncodeToString(peerBuf[48:]),
		conn:   conn,
	}, nil
}

func (p *PeerConn) Receive() (*Payload, error) {
	var length uint32
	err := binary.Read(p.conn, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("binary read length, err = %w", err)
	}

	var messageId byte
	err = binary.Read(p.conn, binary.BigEndian, &messageId)
	if err != nil {
		return nil, fmt.Errorf("binary read message id, err = %w", err)
	}

	if length > 1 {
		data := make([]byte, length-1)
		n, err := io.ReadFull(p.conn, data)
		if err != nil {
			return nil, fmt.Errorf("ReadFull err = %w", err)
		}
		if n != int(length-1) {
			return nil, fmt.Errorf("wanted to read %d, got %d bytes", length-1, n)
		}

		return &Payload{
			Len:       length,
			MessageId: MessageId(messageId),
			Data:      data,
		}, nil
	}

	return &Payload{
		Len:       length,
		MessageId: MessageId(messageId),
		Data:      make([]byte, 0),
	}, nil
}

func (p *PeerConn) Send(messageId MessageId, data []byte) error {
	buf := make([]byte, 4+1+len(data))

	binary.BigEndian.PutUint32(buf[0:4], uint32(len(data)+1))
	buf[4] = byte(messageId)
	copy(buf[5:], data)

	_, err := p.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("conn write, err = %w", err)
	}

	return nil
}

func (p *PeerConn) Download(meta *MetaInfoFile, pieceIndex int) ([]byte, error) {
	length := meta.Info.Length
	// TODO:
	pieceLength := meta.Info.PieceLength
	if pieceIndex >= length/pieceLength {
		pieceLength = length - (pieceLength * pieceIndex)
	}

	blocks := (pieceLength / BlockSize) + 1
	lastBlockSize := pieceLength % BlockSize

	// Perfectly aligned blocks
	if lastBlockSize == 0 {
		blocks--
		lastBlockSize = BlockSize
	}

	pieceHash := meta.Info.Pieces[pieceIndex]
	piece := make([]byte, pieceLength)
	byteOffset := 0

	for block := 0; block < blocks; block++ {
		requestBuffer := make([]byte, 12)

		binary.BigEndian.PutUint32(requestBuffer[0:4], uint32(pieceIndex))
		binary.BigEndian.PutUint32(requestBuffer[4:8], uint32(byteOffset))
		if block == blocks-1 {
			log.Printf("Sending index: %d, offset: %d, last block size: %d\n", pieceIndex, byteOffset, lastBlockSize)
			binary.BigEndian.PutUint32(requestBuffer[8:12], uint32(lastBlockSize))
		} else {
			log.Printf("Sending index: %d, offset: %d, block size: %d\n", pieceIndex, byteOffset, BlockSize)
			binary.BigEndian.PutUint32(requestBuffer[8:12], uint32(BlockSize))
		}

		err := p.Send(RequestMessageId, requestBuffer)
		if err != nil {
			return nil, err
		}

		response, err := p.Receive()
		if err != nil {
			return nil, err
		}

		if response.MessageId != PieceMessageId {
			return nil, fmt.Errorf("invalid message id for a piece: %v", response.MessageId)
		}

		respIndex := binary.BigEndian.Uint32(response.Data[0:4])
		respBegin := binary.BigEndian.Uint32(response.Data[4:8])
		data := response.Data[8:]

		log.Printf("Got piece Index: %d, Begin: %d, Data Size: %d\n", respIndex, respBegin, len(data))

		copy(piece[respBegin:], data)

		byteOffset += BlockSize
	}

	sumHash := sha1.Sum(piece)
	log.Printf("Gathered piece hash: %s, piece hash: %s\n", hex.EncodeToString(sumHash[:]), pieceHash)

	// TODO:
	for i, shas := range meta.Info.Pieces {
		log.Printf("#%d - sha: %s", i, shas)
	}

	return piece, nil
}
