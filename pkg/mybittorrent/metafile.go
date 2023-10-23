package mybittorrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
)

type MetaInfoFile struct {
	Raw      map[string]interface{}
	Announce string
	Info     Info
}

type Info struct {
	Name        string
	Length      int
	PieceLength int
	Pieces      []string
}

func (m MetaInfoFile) HashInfo() []byte {
	info := m.Raw["info"].(map[string]interface{})
	infoHash := sha1.Sum([]byte(Encode(info)))
	return infoHash[:]
}

func ParseMetaInfoFile(filepath string) (*MetaInfoFile, error) {
	fileContent, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	decoded, err := NewDecoder(string(fileContent)).Decode()
	if err != nil {
		return nil, err
	}

	if meta, ok := decoded.(map[string]interface{}); ok {
		info := meta["info"].(map[string]interface{})

		pieces := info["pieces"].(string)
		if len(pieces)%20 != 0 {
			fmt.Println("Pieces not a multiple of 20")
			os.Exit(1)
		}
		buf := bytes.NewBufferString(pieces)
		piecesHashes := make([]string, 0)
		for len(buf.Bytes()) > 0 {
			hash := buf.Next(20)
			piecesHashes = append(piecesHashes, hex.EncodeToString(hash))
		}
		sort.Sort(sort.StringSlice(piecesHashes))

		return &MetaInfoFile{
			Raw:      meta,
			Announce: meta["announce"].(string),
			Info: Info{
				Name:        info["name"].(string),
				Length:      info["length"].(int),
				PieceLength: info["piece length"].(int),
				Pieces:      piecesHashes,
			},
		}, nil
	}

	return nil, fmt.Errorf("metafile not a dict")
}
