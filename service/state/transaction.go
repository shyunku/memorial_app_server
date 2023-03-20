package state

import (
	"crypto/sha256"
	"errors"
	"memorial_app_server/libs/json"
	"reflect"
)

var (
	ErrInvalidTxType = errors.New("invalid transaction type")
	ErrInvalidTxFrom = errors.New("invalid transaction from")
	ErrInvalidTxTime = errors.New("invalid transaction time")
)

type Hash [32]byte

type Transaction struct {
	From      string
	Type      int64
	Timestamp int64
	Content   []byte
}

func NewTransaction(from string, txType int64, timestamp int64, content []byte) *Transaction {
	return &Transaction{
		From:      from,
		Type:      txType,
		Timestamp: timestamp,
		Content:   content,
	}
}

func (tx *Transaction) Hash() Hash {
	fieldBytes := make([]byte, 0)
	values := reflect.ValueOf(tx)

	for i := 0; i < values.NumField(); i++ {
		value := values.Field(i)
		raw := value.Interface()
		bytes, _ := json.Marshal(raw)
		fieldBytes = append(fieldBytes, bytes...)
	}

	hash := sha256.Sum256(fieldBytes)
	return hash
}

func (tx *Transaction) Validate() error {
	// type validation
	if tx.Type < 0 || tx.Type > 1 {
		return ErrInvalidTxType
	}
	// from validation
	if tx.From == "" {
		return ErrInvalidTxFrom
	}
	// timestamp validation
	if tx.Timestamp < 0 {
		return ErrInvalidTxTime
	}
	return nil
}
