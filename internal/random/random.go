package random

import (
	"math/rand"
	"sync"
	"time"
)

const NameSuffixLength = 6

var (
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu     sync.Mutex
)

const lowerAlphaNumeric = "abcdefghijklmnopqrstuvwxyz0123456789"

func AlphaNumericString(length int) string {
	mu.Lock()
	defer mu.Unlock()

	result := make([]uint8, length)
	for i := range result {
		result[i] = lowerAlphaNumeric[random.Intn(len(lowerAlphaNumeric))]
	}
	return string(result)
}
