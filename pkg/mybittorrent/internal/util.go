package internal

import "crypto/rand"

func RandomPeerId() ([]byte, error) {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	return b, err
}
