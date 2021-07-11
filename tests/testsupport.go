package tests

import (
	"math/rand"
	"time"
)

func RandomString() string {
	return RandomStringWithLen(10)
}

func RandomStringWithLen(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))] //nolint: gosec
	}
	return string(s)
}

func RandomInt() int {
	return RandomIntWithMax(10)
}

func RandomIntWithMax(max int) int {
	return rand.Intn(max) //nolint: gosec
}

func RandomDuration() time.Duration {
	return time.Second * time.Duration(RandomInt())
}
