package randomtext

import (
	"math/rand"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func String(n int) string {
	result := make([]byte, n)

	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
