package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"memorial_app_server/log"
	"memorial_app_server/service/database"
	"sync"
	"time"
)

type Chain struct {
	blocks          map[int64]*Block
	lastBlockNumber int64

	lock *sync.Mutex
}

func newStateChain() *Chain {
	// TODO :: add cleaner as go-routine for database
	c := &Chain{
		blocks:          make(map[int64]*Block),
		lastBlockNumber: 0,
		lock:            &sync.Mutex{},
	}
	c.blocks[0] = NewBlock(0, NewState(), nil, Hash{})
	return c
}

func (c *Chain) GetLastBlockNumber() int64 {
	return c.lastBlockNumber
}

func (c *Chain) GetWaitingBlockNumber() int64 {
	return c.lastBlockNumber + 1
}

func (c *Chain) GetBlockHash(number int64) (Hash, error) {
	block, err := c.GetBlockByNumber(number)
	if err != nil {
		return Hash{}, err
	}
	return block.Hash(), nil
}

func (c *Chain) GetBlocksByInterval(start, end int64) ([]*Block, error) {
	var blocks []*Block
	for i := start; i <= end; i++ {
		block, err := c.GetBlockByNumber(i)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (c *Chain) GetBlockByNumber(number int64) (*Block, error) {
	if number < 0 {
		return nil, fmt.Errorf("invalid block number: %d", number)
	}

	// find block on cache
	block, exist := c.blocks[number]
	if !exist {
		// find block on database
		var blockEntity database.BlockEntity
		err := database.DB.Get(&blockEntity, "SELECT * FROM blocks WHERE block_number = ?", number)
		if err != nil {
			log.Error(err)
			log.Debug("block number: ", number)
			return nil, err
		}

		state := NewState()
		err = state.FromBytes(blockEntity.State)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		var txEntity database.TransactionEntity
		err = database.DB.Get(&txEntity, "SELECT * FROM transactions WHERE hash = ?", blockEntity.TxHash)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// get previous blockHash
		prevBlockHash, err := c.GetBlockHash(number - 1)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		tx := NewTransaction(*txEntity.From, *txEntity.Type, *txEntity.Timestamp, txEntity.Content)
		block = NewBlock(number, state, tx, prevBlockHash)
		c.InsertBlock(block)
	}

	return block, nil
}

func (c *Chain) GetBlockByHash(hash Hash) (*Block, error) {
	// TODO :: optimize this (maybe use cache)
	var blockEntity database.BlockEntity
	err := database.DB.Get(&blockEntity, "SELECT * FROM blocks WHERE block_hash = ?", hash)
	if err != nil {
		return nil, err
	}

	state := NewState()
	err = state.FromBytes(blockEntity.State)
	if err != nil {
		return nil, err
	}

	var txEntity database.TransactionEntity
	err = database.DB.Get(&txEntity, "SELECT * FROM transactions WHERE hash = ?", blockEntity.TxHash)
	if err != nil {
		return nil, err
	}

	prevBlockHash, err := hexToHash(*blockEntity.PrevBlockHash)
	if err != nil {
		return nil, err
	}

	tx := NewTransaction(*txEntity.From, *txEntity.Type, *txEntity.Timestamp, txEntity.Content)
	block := NewBlock(*blockEntity.Number, state, tx, prevBlockHash)

	return block, nil
}

func (c *Chain) InsertBlock(block *Block) {
	if _, ok := c.blocks[block.Number]; ok {
		// already exists
		return
	}
	c.blocks[block.Number] = block
	if block.Number > c.lastBlockNumber {
		c.lastBlockNumber = block.Number
	}
}

// ApplyTransaction applies a transaction to build the new state
func (c *Chain) ApplyTransaction(tx *Transaction) (*Block, error) {
	var lastState *State
	c.lock.Lock()

	lastBlock, exists := c.blocks[c.lastBlockNumber]
	if !exists {
		lastState = NewState()
	} else {
		lastState = lastBlock.State
	}

	if lastState == nil {
		return nil, errors.New("last state is nil")
	}

	// apply transaction
	newState, err := ExecuteTransaction(lastState, tx)
	if err != nil {
		return nil, err
	}

	// create new block
	newBlockNumber := c.lastBlockNumber + 1
	newBlock := NewBlock(newBlockNumber, newState, tx, lastBlock.Hash())

	// update chain
	c.InsertBlock(newBlock)
	c.lock.Unlock()

	// save block & transaction to database
	go func() {
		var marshaledContent []byte
		if tx.Content != nil {
			marshaledContent, err = json.Marshal(tx.Content)
			if err != nil {
				log.Error(err)
				return
			}
		}

		marshaledState, err := newState.ToBytes()
		if err != nil {
			log.Error(err)
			return
		}

		ctx, err := database.DB.BeginTxx(context.Background(), nil)
		if err != nil {
			log.Error(err)
			return
		}

		dateTime := time.Unix(0, tx.Timestamp*int64(time.Millisecond))
		_, err = ctx.Exec("INSERT INTO transactions (type, `from`, timestamp, content, hash) VALUES (?, ?, ?, ?, ?)", tx.Type, tx.From, dateTime, marshaledContent, tx.Hash().Hex())
		if err != nil {
			log.Error(err)
			err := ctx.Rollback()
			if err != nil {
				return
			}
			return
		}

		_, err = database.DB.Exec(
			"INSERT INTO blocks (state, block_number, block_hash, tx_hash) VALUES (?, ?, ?, ?)",
			marshaledState, newBlock.Number, newBlock.Hash().Hex(), newBlock.Tx.Hash().Hex(),
		)
		if err != nil {
			log.Error(err)
			err := ctx.Rollback()
			if err != nil {
				log.Error(err)
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

	return newBlock, nil
}
