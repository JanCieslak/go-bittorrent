package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strconv"
	"unicode"
)

func main() {
	if len(os.Args) <= 2 {
		fmt.Println("Command not specified")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "decode":
		if len(os.Args) != 3 {
			fmt.Println("No value to decode specified")
			os.Exit(1)
		}

		decoded, err := NewDecoder(os.Args[2]).Decode()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		jsonOutput, err := json.Marshal(decoded)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	case "info":
		if len(os.Args) != 3 {
			fmt.Println("No torrent file to parse specified")
			os.Exit(1)
		}

		// TODO: Should return Meta struct
		meta, err := ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Tracker URL:", meta.Announce.String())
		fmt.Println("Length:", meta.Info.Length)
		fmt.Println("Info Hash:", meta.HashInfo())
		fmt.Println("Piece Length:", meta.Info.PieceLength)
		fmt.Println("Piece Hashes:")
		for _, p := range meta.Info.Pieces {
			fmt.Println(p)
		}
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

// Bencode library

type MetaInfoFile struct {
	Raw      Dictionary
	Announce url.URL
	Info     Info
}

type Info struct {
	Name        string
	Length      int
	PieceLength int
	Pieces      []string
}

func (m MetaInfoFile) HashInfo() string {
	info := m.Raw["info"].(Dictionary)
	return info.HashInfo()
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

	if meta, ok := decoded.(Dictionary); ok {
		announce, err := url.Parse(meta["announce"].(string))
		if err != nil {
			return nil, err
		}

		info := meta["info"].(Dictionary)

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
			Announce: *announce,
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

// Decoder

type Dictionary map[string]interface{}

func (d Dictionary) EncodeInfo() string {
	buf := new(bytes.Buffer)
	buf.WriteRune('d')
	d.inSortedOrder(func(k string, v interface{}) {
		_, _ = buf.WriteString(Encode(k))
		_, _ = buf.WriteString(Encode(v))
	})
	buf.WriteRune('e')
	return buf.String()
}

func (d Dictionary) HashInfo() string {
	encodedDict := d.EncodeInfo()
	log.Println(encodedDict)
	shaSum := sha1.Sum([]byte(encodedDict))
	return hex.EncodeToString(shaSum[:])
}

func (d Dictionary) inSortedOrder(fn func(k string, v interface{})) {
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fn(key, d[key])
	}
}

type Decoder struct {
	*bytes.Buffer
}

func NewDecoder(encoded string) *Decoder {
	return &Decoder{
		Buffer: bytes.NewBufferString(encoded),
	}
}

func (b *Decoder) Decode() (interface{}, error) {
	return b.decodeBencode()
}

func (b *Decoder) decodeBencode() (interface{}, error) {
	prefix, _, err := b.ReadRune()
	if err != nil {
		return nil, err
	}
	switch {
	case unicode.IsDigit(prefix):
		err = b.UnreadRune()
		if err != nil {
			return nil, err
		}
		return b.decodeString()
	case prefix == 'i':
		return b.decodeInteger()
	case prefix == 'l':
		return b.decodeList()
	case prefix == 'd':
		return b.decodeDict()
	default:
		return "", fmt.Errorf("prefix not recognized: %s", string(prefix))
	}
}

func (b *Decoder) decodeString() (string, error) {
	strLen, err := b.ReadString(':')
	if err != nil {
		return "", err
	}
	strLen = strLen[:len(strLen)-1]

	intStrLen, err := strconv.Atoi(strLen)
	if err != nil {
		return "", err
	}

	return string(b.Next(intStrLen)), nil
}

func (b *Decoder) decodeInteger() (int, error) {
	num, err := b.ReadString('e')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(num[:len(num)-1])
}

func (b *Decoder) decodeList() ([]interface{}, error) {
	list := make([]interface{}, 0)

	nextRune, _, err := b.ReadRune()
	if err != nil {
		return nil, err
	}

	for nextRune != 'e' {
		err = b.UnreadRune()
		if err != nil {
			return nil, err
		}

		item, err := b.Decode()
		if err != nil {
			return nil, err
		}

		list = append(list, item)

		nextRune, _, err = b.ReadRune()
		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (b *Decoder) decodeDict() (Dictionary, error) {
	dict := make(Dictionary)

	nextRune, _, err := b.ReadRune()
	if err != nil {
		return nil, err
	}

	for nextRune != 'e' {
		err = b.UnreadRune()
		if err != nil {
			return nil, err
		}

		key, err := b.decodeString()
		if err != nil {
			return nil, err
		}

		value, err := b.Decode()
		if err != nil {
			return nil, err
		}

		dict[key] = value

		nextRune, _, err = b.ReadRune()
		if err != nil {
			return nil, err
		}
	}

	return dict, nil
}

func Encode(v interface{}) string {
	switch value := v.(type) {
	case string:
		return fmt.Sprintf("%d:%s", len(value), value)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("i%de", value)
	case []interface{}:
		buf := new(bytes.Buffer)
		for _, item := range value {
			buf.WriteString(Encode(item))
		}
		return fmt.Sprintf("l%se", buf.String())
	case Dictionary:
		buf := new(bytes.Buffer)
		for key, value2 := range value {
			log.Println("key:", key)
			buf.WriteString(Encode(key))
			buf.WriteString(Encode(value2))
		}
		return fmt.Sprintf("d%se", buf.String())
	default:
		log.Println("unsupported case of", value)
		os.Exit(1)
		return ""
	}
}
