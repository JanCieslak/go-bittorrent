package bencoding

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

func DecodeString(encoded string) (interface{}, error) {
	return newDecoder(bytes.NewBufferString(encoded)).decode()
}

func Decode(encoded []byte) (interface{}, error) {
	return newDecoder(bytes.NewBuffer(encoded)).decode()
}

type decoder struct {
	*bytes.Buffer
}

func newDecoder(b *bytes.Buffer) *decoder {
	return &decoder{
		Buffer: b,
	}
}

func (b *decoder) decode() (interface{}, error) {
	value, err := b.decodeBencode()
	if err != nil {
		return nil, fmt.Errorf("unable to decode bencoded value, err = %w", err)
	}
	return value, nil
}

func (b *decoder) decodeBencode() (interface{}, error) {
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

func (b *decoder) decodeString() (string, error) {
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

func (b *decoder) decodeInteger() (int, error) {
	num, err := b.ReadString('e')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(num[:len(num)-1])
}

func (b *decoder) decodeList() ([]interface{}, error) {
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

		item, err := b.decodeBencode()
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

func (b *decoder) decodeDict() (map[string]interface{}, error) {
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

		value, err := b.decodeBencode()
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
