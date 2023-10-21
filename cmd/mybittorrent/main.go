package main

import (
	"encoding/json"
	"log"
	"strings"
	// Uncomment this line to pass the first stage
	// "encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
	switch prefix := rune(bencodedString[0]); {
	case unicode.IsDigit(prefix):
		return decodeString(bencodedString)
	case prefix == 'i':
		return decodeInteger(bencodedString)
	default:
		return "", fmt.Errorf("could not decode: %s", bencodedString)
	}
}

func decodeString(bencodedString string) (string, error) {
	s := strings.SplitN(bencodedString, ":", 2)
	if len(s) != 2 {
		return "", fmt.Errorf("invalid bencoded string %s", bencodedString)
	}

	strLen := s[0]
	str := s[1]

	length, err := strconv.Atoi(strLen)
	if err != nil {
		return "", err
	}

	return str[:length], nil
}

func decodeInteger(bencodedString string) (int, error) {
	return strconv.Atoi(bencodedString[1 : len(bencodedString)-1])
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	log.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
