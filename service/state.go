package service

import (
	"crypto/sha256"
	"math/big"
)

type StateHash [32]byte

var InitialStateHash = StateHash{}

type State struct {
	Index *big.Int
	Hash  StateHash
}

func InitialState() *State {
	return &State{
		Index: big.NewInt(0),
		Hash:  InitialStateHash,
	}
}

func Validate() error {
	return nil
}

func (s *State) NextHash(tx *Transaction) StateHash {
	prevHash := s.Hash
	txHash := tx.Hash()
	nextHash := sha256.Sum256(append(prevHash[:], txHash[:]...))
	return nextHash
}
