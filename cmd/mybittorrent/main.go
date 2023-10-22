package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

		meta, err := ParseMetaInfoFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		info := meta["info"].(Dictionary)
		sum := sha1.Sum([]byte(Encode(info)))
		fmt.Println("Tracker URL:", meta["announce"])
		fmt.Println("Length:", info["length"])
		fmt.Println("Info Hash:", hex.EncodeToString(sum[:]))
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

// Bencode library

func ParseMetaInfoFile(filepath string) (Dictionary, error) {
	fileContent, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	decoded, err := NewDecoder(string(fileContent)).Decode()
	if err != nil {
		return nil, err
	}

	if meta, ok := decoded.(Dictionary); ok {
		return meta, nil
	}

	return nil, fmt.Errorf("metafile not a dict")
}

// Decoder

type Dictionary map[string]interface{}

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
