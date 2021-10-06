package test

import (
	"fmt"
	"math/rand"
	"time"
)

func randomName(prefix string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s%04x", prefix, r.Uint32()>>16)
}
