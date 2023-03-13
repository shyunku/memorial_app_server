package service

import "math/big"

type Syncer struct {
	userId     string
	stateCache map[*big.Int]*State
	lastState  *State
}

func NewSyncer(userId string) *Syncer {
	// TODO :: load from database
	return &Syncer{
		userId:    userId,
		lastState: InitialState(),
	}
}

// ApplyTransaction applies a transaction to the current state and returns the new state.
func (s *State) ApplyTransaction(tx *Transaction) (*State, error) {
	nextHash := s.NextHash(tx)

	// apply transaction
	if err := ExecuteTransaction(tx); err != nil {
		return nil, err
	}

	return &State{
		Index: big.NewInt(0).Add(s.Index, big.NewInt(1)),
		Hash:  nextHash,
	}, nil
}
