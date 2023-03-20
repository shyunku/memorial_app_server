package state

import (
	"context"
	"errors"
	"math/big"
	"memorial_app_server/log"
	"memorial_app_server/service/database"
	"sync"
)

type Chain struct {
	blocks          map[*big.Int]*Block
	lastBlockNumber *big.Int

	lock *sync.Mutex
}

func newStateChain() *Chain {
	// TODO :: add cleaner as go-routine for database
	return &Chain{
		blocks:          make(map[*big.Int]*Block),
		lastBlockNumber: big.NewInt(0),
	}
}

func (c *Chain) GetLastBlockNumber() *big.Int {
	return c.lastBlockNumber
}

func (c *Chain) GetWaitingBlockNumber() *big.Int {
	return big.NewInt(0).Add(c.lastBlockNumber, big.NewInt(1))
}

func (c *Chain) InsertBlock(block *Block) {
	if _, ok := c.blocks[block.Index]; ok {
		// already exists
		return
	}
	c.blocks[block.Index] = block
	if block.Index.Cmp(c.lastBlockNumber) > 0 {
		c.lastBlockNumber = block.Index
	}
}

// ApplyTransaction applies a transaction to build the new state
func (c *Chain) ApplyTransaction(tx *Transaction) error {
	var lastState *State
	c.lock.Lock()

	lastBlock, exists := c.blocks[c.lastBlockNumber]
	if !exists {
		lastState = NewState()
	} else {
		lastState = lastBlock.State
	}

	if lastState == nil {
		return errors.New("last state is nil")
	}

	// apply transaction
	newState, err := ExecuteTransaction(lastState, tx)
	if err != nil {
		return err
	}

	// create new block
	newBlockNumber := big.NewInt(0).Add(c.lastBlockNumber, big.NewInt(1))
	newBlock := NewStateBlock(newBlockNumber, newState, tx)

	// update chain
	c.InsertBlock(newBlock)
	c.lock.Unlock()

	// save block & transaction to database
	go func() {
		ctx, err := database.DB.BeginTxx(context.Background(), nil)
		if err != nil {
			log.Error(err)
			return
		}

		_, err = ctx.Exec("INSERT INTO transactions (type, from, timestamp, content, hash) VALUES (?, ?, ?, ?, ?)", tx.Type, tx.From, tx.Timestamp, tx.Content, tx.Hash())
		if err != nil {
			log.Error(err)
			err := ctx.Rollback()
			if err != nil {
				return
			}
			return
		}

		_, err = database.DB.Exec("INSERT INTO blocks (uid, state, block_number, tx_hash) VALUES (?, ?, ?, ?)", tx.From, newBlock.State, newBlock.Index, newBlock.Tx.Hash())
		if err != nil {
			log.Error(err)
			err := ctx.Rollback()
			if err != nil {
				return
			}
			return
		}

		err = ctx.Commit()
		if err != nil {
			log.Error(err)
			return
		}

		log.Debug("transaction/block saved successfully")
	}()

	return nil
}
