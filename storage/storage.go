package storage

import (
	"math/rand"
	"time"
)

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
}

// GenerateID allows us to easily generate a new ID. If we want to
// make a new userId, we can call storage.GenerateID("USR_", 15) or something of the like
func GenerateID(prefix string, length int) string {
	rand.Seed(time.Now().UnixNano())
	base := randStringRunes(length)
	return prefix + base
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}
