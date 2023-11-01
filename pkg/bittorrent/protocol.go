package bittorrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

const ProtocolHeader = "BitTorrent protocol"

type MessageId byte

const (
	UnchokeMessageId    MessageId = 1
	InterestedMessageId MessageId = 2
	BitFieldMessageId   MessageId = 5
	RequestMessageId    MessageId = 6
	PieceMessageId      MessageId = 7

	BlockSize = 16 * 1024
)

type RequestMessageData struct {
	PieceIndex         uint32
	ByteOffset         uint32
	RequestedBlockSize uint32
}

type PieceMessageData struct {
	Index  uint32
	Offset uint32
	Data   []byte
}

type Payload struct {
	Len       uint32
	MessageId MessageId
	Data      []byte
}

type Conn struct {
	PeerID string
	conn   *net.TCPConn
}

func Dial(meta *MetaInfoFile, address string) (*Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	peerId, err := randomPeerId()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.WriteByte(19)
	buf.WriteString(ProtocolHeader)
	buf.Write(make([]byte, 8))
	buf.Write(meta.HashInfo())
	buf.Write(peerId)

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	peerBuf := make([]byte, 68)
	_, err = conn.Read(peerBuf)
	if err != nil {
		return nil, err
	}

	return &Conn{
		PeerID: hex.EncodeToString(peerBuf[48:]),
		conn:   conn,
	}, nil
}

func (c *Conn) Receive() (*Payload, error) {
	var length uint32
	err := binary.Read(c.conn, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("binary read length, err = %w", err)
	}

	var messageId byte
	err = binary.Read(c.conn, binary.BigEndian, &messageId)
	if err != nil {
		return nil, fmt.Errorf("binary read message id, err = %w", err)
	}

	var data []byte
	if length > 1 {
		data = make([]byte, length-1)
		n, err := io.ReadFull(c.conn, data)
		if err != nil {
			return nil, fmt.Errorf("ReadFull err = %w", err)
		}
		if n != int(length-1) {
			return nil, fmt.Errorf("wanted to read %d bytes, got %d bytes", length-1, n)
		}
	}

	return &Payload{
		Len:       length,
		MessageId: MessageId(messageId),
		Data:      data,
	}, nil
}

func (c *Conn) Send(messageId MessageId, data []byte) error {
	buf := make([]byte, 4+1+len(data))

	binary.BigEndian.PutUint32(buf[0:4], uint32(len(data)+1))
	buf[4] = byte(messageId)
	copy(buf[5:], data)

	_, err := c.conn.Write(buf)
	if err != nil {
		return fmt.Errorf("send error, err = %w", err)
	}

	return nil
}

func (c *Conn) Download(meta *MetaInfoFile) ([]byte, error) {
	buf := new(bytes.Buffer)

	for i, _ := range meta.Info.Pieces {
		data, err := c.DownloadPiece(meta, i)
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(data)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (c *Conn) DownloadPiece(meta *MetaInfoFile, pieceIndex int) ([]byte, error) {
	// Calculate piece length
	pieceLength := meta.Info.PieceLength
	isLastPiece := pieceIndex >= (meta.Info.Length / meta.Info.PieceLength)
	if isLastPiece {
		pieceLength = meta.Info.Length - (pieceLength * pieceIndex)
	}

	numberOfBlocks := (pieceLength / BlockSize) + 1
	lastBlockSize := pieceLength % BlockSize

	// Perfectly aligned blocks
	if lastBlockSize == 0 {
		numberOfBlocks--
		lastBlockSize = BlockSize
	}

	pieceData := make([]byte, pieceLength)
	byteOffset := 0

	for blockIndex := 0; blockIndex < numberOfBlocks; blockIndex++ {
		isLastBlock := blockIndex == numberOfBlocks-1
		requestedBlockSize := BlockSize
		if isLastBlock {
			requestedBlockSize = lastBlockSize
		}

		err := c.sendRequestMessage(RequestMessageData{
			PieceIndex:         uint32(pieceIndex),
			ByteOffset:         uint32(byteOffset),
			RequestedBlockSize: uint32(requestedBlockSize),
		})
		if err != nil {
			return nil, err
		}

		response, err := c.receivePieceMessage()
		if err != nil {
			return nil, err
		}

		if uint32(byteOffset) != response.Offset {
			return nil, fmt.Errorf("invalid bytes offset, requested: %d, got: %d", byteOffset, response.Offset)
		}

		if uint32(pieceIndex) != response.Index {
			return nil, fmt.Errorf("invalid piece index, requested: %d, got: %d", pieceIndex, response.Index)
		}

		copy(pieceData[response.Offset:], response.Data)
		byteOffset += BlockSize
	}

	expectedPieceHash := meta.Info.Pieces[pieceIndex]
	sumHash := sha1.Sum(pieceData)
	actualPieceHash := hex.EncodeToString(sumHash[:])

	log.Println(meta.Info.Pieces)

	if expectedPieceHash != actualPieceHash {
		return nil, fmt.Errorf("invalid data, expected hash: %s, got: %s", expectedPieceHash, actualPieceHash)
	}

	return pieceData, nil
}

func (c *Conn) sendRequestMessage(request RequestMessageData) error {
	requestBuffer := make([]byte, 12)

	binary.BigEndian.PutUint32(requestBuffer[0:4], request.PieceIndex)
	binary.BigEndian.PutUint32(requestBuffer[4:8], request.ByteOffset)
	binary.BigEndian.PutUint32(requestBuffer[8:12], request.RequestedBlockSize)

	err := c.Send(RequestMessageId, requestBuffer)
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) receivePieceMessage() (*PieceMessageData, error) {
	response, err := c.Receive()
	if err != nil {
		return nil, err
	}

	if response.MessageId != PieceMessageId {
		return nil, fmt.Errorf("invalid message id, expected: %v, got: %v", PieceMessageId, response.MessageId)
	}

	return &PieceMessageData{
		Index:  binary.BigEndian.Uint32(response.Data[0:4]),
		Offset: binary.BigEndian.Uint32(response.Data[4:8]),
		Data:   response.Data[8:],
	}, nil
}
