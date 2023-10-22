package mybittorrent

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

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

func (b *Decoder) decodeDict() (map[string]interface{}, error) {
	dict := make(map[string]interface{})

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
