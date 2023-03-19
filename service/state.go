package service

import (
	"crypto/sha256"
	"math/big"
)

type StateHash [32]byte

var InitialStateHash = StateHash{}

type State struct {
}

type StateIndexer struct {
	Index *big.Int
	Hash  StateHash
	State *State
}

func InitialStateIndexer() *StateIndexer {
	return &StateIndexer{
		Index: big.NewInt(0),
		Hash:  InitialStateHash,
	}
}

func Validate() error {
	return nil
}

func (s *StateIndexer) NextHash(tx *Transaction) StateHash {
	prevHash := s.Hash
	txHash := tx.Hash()
	nextHash := sha256.Sum256(append(prevHash[:], txHash[:]...))
	return nextHash
}
