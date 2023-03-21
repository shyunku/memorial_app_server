package state

import (
	"crypto/sha256"
	"math/big"
)

type Block struct {
	Index         *big.Int
	State         *State
	Tx            *Transaction // transaction that is currently being applied
	PrevBlockHash Hash
}

func NewBlock(index *big.Int, state *State, tx *Transaction, prevBlockHash Hash) *Block {
	return &Block{
		Index:         index,
		State:         state,
		Tx:            tx,
		PrevBlockHash: prevBlockHash,
	}
}

func (b *Block) Hash() Hash {
	var bytes []byte
	bytes = append(bytes, b.Index.Bytes()...)
	bytes = append(bytes, b.State.Hash().Bytes()...)
	bytes = append(bytes, b.Tx.Hash().Bytes()...)
	bytes = append(bytes, b.PrevBlockHash.Bytes()...)
	hash := sha256.Sum256(bytes)
	return hash
}
