package service

import (
	"crypto/sha256"
	"errors"
	"memorial_app_server/libs/json"
	"reflect"
)

var (
	ErrInvalidTxType = errors.New("invalid transaction type")
)

type Transaction struct {
	UserId  string
	Type    int64
	Content any
}

func (tx *Transaction) Hash() StateHash {
	fieldBytes := make([]byte, 0)
	values := reflect.ValueOf(tx)

	for i := 0; i < values.NumField(); i++ {
		value := values.Field(i)
		raw := value.Interface()
		bytes, _ := json.Marshal(raw)
		fieldBytes = append(fieldBytes, bytes...)
	}

	hash := sha256.Sum256(fieldBytes)
	return StateHash(hash)
}
