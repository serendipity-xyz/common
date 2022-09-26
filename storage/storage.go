package storage

import (
	"math/rand"
	"time"

	"github.com/serendipity-xyz/common/types"
)

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// Decoder represents an object that can be decoded (unmarshalled) into
// an interface object
type Decoder interface {
	Decode(v interface{}) error
}

// Manager represents a struct that can interface with a backing data store
type Manager interface {
	FindOne(l types.Logger, cc *CallContext, params *FindOneParams) (Decoder, error)
	FindMany(l types.Logger, cc *CallContext, params *FindManyParams) (Decoder, error)
	InsertOne(l types.Logger, cc *CallContext, document interface{}, params *InsertOneParams) (interface{}, error)
	InsertMany(l types.Logger, cc *CallContext, data []interface{}, params *InsertManyParams) (interface{}, error)
	Upsert(l types.Logger, cc *CallContext, updates interface{}, params *UpsertParams) (int64, error)
	Delete(l types.Logger, cc *CallContext, params *DeleteParams) (int64, error)
	Close(l types.Logger)
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
