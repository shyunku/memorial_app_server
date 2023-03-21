package state

import (
	"context"
	"errors"
	"math/big"
	"memorial_app_server/log"
	"memorial_app_server/service/database"
	"memorial_app_server/util"
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

func (c *Chain) GetBlockHash(number *big.Int) (Hash, error) {
	block, err := c.GetBlockByNumber(number)
	if err != nil {
		return Hash{}, err
	}
	return block.Hash(), nil
}

func (c *Chain) GetBlocksByInterval(start, end *big.Int) ([]*Block, error) {
	var blocks []*Block
	for i := new(big.Int).Set(start); i.Cmp(end) <= 0; i.Add(i, util.Big1) {
		block, err := c.GetBlockByNumber(i)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (c *Chain) GetBlockByNumber(number *big.Int) (*Block, error) {
	// find block on cache
	block, exist := c.blocks[number]
	if !exist {
		// find block on database
		var blockEntity database.BlockEntity
		err := database.DB.Get(&blockEntity, "SELECT * FROM blocks WHERE block_number = ?", number)
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

		// get previous blockHash
		prevBlockHash, err := c.GetBlockHash(big.NewInt(0).Sub(number, util.Big1))
		if err != nil {
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
	block := NewBlock(big.NewInt(*blockEntity.Number), state, tx, prevBlockHash)

	return block, nil
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

	// get previous block
	prevBlock, err := c.GetBlockByNumber(big.NewInt(0).Sub(c.lastBlockNumber, util.Big1))
	if err != nil {
		return nil, err
	}

	// create new block
	newBlockNumber := big.NewInt(0).Add(c.lastBlockNumber, big.NewInt(1))
	newBlock := NewBlock(newBlockNumber, newState, tx, prevBlock.Hash())

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

	return newBlock, nil
}
