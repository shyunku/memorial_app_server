package state

import (
	"crypto/sha256"
	"encoding/binary"
)

type Block struct {
	Number        int64
	State         *State
	Tx            *Transaction // transaction that is currently being applied
	PrevBlockHash Hash
}

func NewBlock(number int64, state *State, tx *Transaction, prevBlockHash Hash) *Block {
	return &Block{
		Number:        number,
		State:         state,
		Tx:            tx,
		PrevBlockHash: prevBlockHash,
	}
}

func (b *Block) Hash() Hash {
	var bytes []byte

	blockNumberBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockNumberBytes, uint64(b.Number))

	bytes = append(bytes, blockNumberBytes...)
	if b.State != nil {
		bytes = append(bytes, b.State.Hash().Bytes()...)
	}
	if b.Tx != nil {
		bytes = append(bytes, b.Tx.Hash().Bytes()...)
	}
	bytes = append(bytes, b.PrevBlockHash.Bytes()...)
	hash := sha256.Sum256(bytes)
	return hash
}
