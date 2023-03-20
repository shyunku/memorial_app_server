package state

import "math/big"

type Block struct {
	Index *big.Int
	State *State
	Tx    *Transaction // transaction that is currently being applied
}

func NewStateBlock(index *big.Int, state *State, tx *Transaction) *Block {
	return &Block{
		Index: index,
		State: state,
		Tx:    tx,
	}
}
