package storage

import (
	"math/rand"
	"time"
)

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// Decoder represents an object that can be decoded (unmarshalled) into
// an interface object
type Decoder interface {
	Decode(v interface{}) error
}

// Manager represents a struct that can interface with a backing data store
type Manager interface {
	FindOne(l logger, cc *callContext, params *FindOneParams) (Decoder, error)
	FindMany(l logger, cc *callContext, params *FindManyParams) (Decoder, error)
	InsertOne(l logger, cc *callContext, document interface{}, params *InsertOneParams) (interface{}, error)
	InsertMany(l logger, cc *callContext, data interface{}, params *InsertManyParams) (interface{}, error)
	Upsert(l logger, cc *callContext, updates interface{}, params *UpsertParams) (int64, error)
	Delete(l logger, cc *callContext, params *DeleteParams) (int64, error)
	Close(l logger)
}

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
