package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"memorial_app_server/log"
	"memorial_app_server/service/database"
	"sync"
)

type Chain struct {
	Blocks          map[int64]*Block `json:"blocks"`
	LastBlockNumber int64            `json:"last_block_number"`
	lock            *sync.Mutex
}

func newStateChain() *Chain {
	// TODO :: add cleaner as go-routine for database
	c := &Chain{
		Blocks:          make(map[int64]*Block),
		LastBlockNumber: 0,
		lock:            &sync.Mutex{},
	}
	newBlock := NewBlock(0, NewState(), nil, Hash{})
	c.Blocks[0] = newBlock
	return c
}

func (c *Chain) GetLastState() *State {
	return c.Blocks[c.LastBlockNumber].State
}

func (c *Chain) GetLastBlockNumber() int64 {
	return c.LastBlockNumber
}

func (c *Chain) GetWaitingBlockNumber() int64 {
	return c.LastBlockNumber + 1
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
	block, exist := c.Blocks[number]
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

		tx := NewTransaction(*txEntity.Version, *txEntity.From, *txEntity.Type, *txEntity.Timestamp, txEntity.Content)
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

	tx := NewTransaction(*txEntity.Version, *txEntity.From, *txEntity.Type, *txEntity.Timestamp, txEntity.Content)
	block := NewBlock(*blockEntity.Number, state, tx, prevBlockHash)

	return block, nil
}

func (c *Chain) DeleteBlockByInterval(start, end int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// delete in cache
	for i := start; i <= end; i++ {
		delete(c.Blocks, i)
	}

	// delete in database
	_, err := database.DB.Exec("DELETE FROM blocks WHERE block_number >= ? AND block_number <= ?", start, end)
	if err != nil {
		return err
	}

	// check if block exists after end in cache
	remain := 0
	for _, block := range c.Blocks {
		if block.Number > end {
			remain++
		}
	}
	if remain > 0 {
		log.Errorf("remain %d blocks in cache after deleting blocks from %d to %d", remain, start, end)
	}

	// update last block number
	c.LastBlockNumber = start - 1

	return nil
}

func (c *Chain) InsertBlock(block *Block) {
	if _, ok := c.Blocks[block.Number]; ok {
		// already exists
		return
	}
	c.Blocks[block.Number] = block
	if block.Number > c.LastBlockNumber {
		c.LastBlockNumber = block.Number
	}
}

// ApplyTransaction applies a transaction to build the new state
func (c *Chain) ApplyTransaction(tx *Transaction) (*Block, error) {
	log.Debug("Applying transaction...")
	var lastState *State
	c.lock.Lock()
	defer c.lock.Unlock()

	lastBlock, exists := c.Blocks[c.LastBlockNumber]
	if !exists {
		lastState = NewState()
	} else {
		lastState = lastBlock.State
	}

	if lastState == nil {
		return nil, errors.New("last state is nil")
	}

	newBlockNumber := c.LastBlockNumber + 1

	// apply transaction
	newState, err := ExecuteTransaction(lastState, tx, newBlockNumber)
	if err != nil {
		return nil, err
	}

	// validate new state
	if err := newState.Validate(); err != nil {
		return nil, err
	}

	// create new block
	newBlock := NewBlock(newBlockNumber, newState, tx, lastBlock.Hash())

	// update chain
	c.InsertBlock(newBlock)
	log.Debugf("block %d inserted to cache successfully", newBlock.Number)

	// save block & transaction to database
	func() {
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

		_, err = ctx.Exec("INSERT INTO transactions (version, type, `from`, timestamp, content, hash) VALUES (?, ?, ?, ?, ?, ?)", tx.Version, tx.Type, tx.From, tx.Timestamp, marshaledContent, tx.Hash.Hex())
		if err != nil {
			log.Error(err)
			err := ctx.Rollback()
			if err != nil {
				return
			}
			return
		}

		_, err = ctx.Exec(
			"INSERT INTO blocks (state, block_number, block_hash, tx_hash, prev_block_hash) VALUES (?, ?, ?, ?, ?)",
			marshaledState,
			newBlock.Number,
			newBlock.Hash().Hex(),
			newBlock.Tx.Hash.Hex(),
			newBlock.PrevBlockHash.Hex(),
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

		log.Debugf("transaction/block %d saved successfully", newBlock.Number)
	}()

	return newBlock, nil
}
