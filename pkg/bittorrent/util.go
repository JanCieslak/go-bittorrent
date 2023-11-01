package bittorrent

import "crypto/rand"

func randomPeerId() ([]byte, error) {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	return b, err
}
