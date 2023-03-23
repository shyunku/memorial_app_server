package state

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"memorial_app_server/libs/json"
	"reflect"
)

var (
	ErrInvalidTxType = errors.New("invalid transaction type")
	ErrInvalidTxFrom = errors.New("invalid transaction from")
	ErrInvalidTxTime = errors.New("invalid transaction time")
	ErrStateMismatch = errors.New("state mismatch")

	SchemeVersion = 0
)

type Hash [32]byte

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

func hexToHash(str string) (Hash, error) {
	// convert hex string to byte array
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return Hash{}, err
	}
	// convert byte array to hash
	var hash Hash
	copy(hash[:], bytes)
	return hash, nil
}

type Transaction struct {
	Version   int64       `json:"version"`
	From      string      `json:"from"`
	Type      int64       `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Content   interface{} `json:"content"`
}

func NewTransaction(from string, txType int64, timestamp int64, content interface{}) *Transaction {
	if SchemeVersion == 0 {
		panic("txType cannot be 0, maybe env is not set correctly")
	}

	return &Transaction{
		Version:   1,
		From:      from,
		Type:      txType,
		Timestamp: timestamp,
		Content:   content,
	}
}

func (tx *Transaction) Hash() Hash {
	fieldBytes := make([]byte, 0)
	values := reflect.ValueOf(*tx)

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
	if tx.Type < 1 || tx.Type > 30 {
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
