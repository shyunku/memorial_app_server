package state

import (
	"github.com/jmoiron/sqlx"
	"math/big"
	"memorial_app_server/log"
	"memorial_app_server/service/database"
)

var Chains *ChainCluster = nil

func InitializeService(db *sqlx.DB) error {
	Chains = NewChainCluster()
	if err := Chains.LoadFromDatabase(db); err != nil {
		return err
	}
	return nil
}

type ChainCluster map[string]*Chain

func NewChainCluster() *ChainCluster {
	return &ChainCluster{}
}

func (sm *ChainCluster) GetChain(userId string) *Chain {
	chain, ok := (*sm)[userId]
	if !ok {
		chain = newStateChain()
		(*sm)[userId] = chain
	}
	return chain
}

func (sm *ChainCluster) LoadFromDatabase(db *sqlx.DB) error {
	// load transactions
	var transactions map[string]*Transaction
	txRows, err := db.Queryx("SELECT * FROM transactions")
	if err != nil {
		return err
	}

	for txRows.Next() {
		var tx database.TransactionEntity
		if err := txRows.StructScan(&tx); err != nil {
			log.Errorf("failed to scan struct of transaction %v ", err)
			return err
		}
		transactions[*tx.Hash] = NewTransaction(
			*tx.From,
			*tx.Type,
			*tx.Timestamp,
			tx.Content,
		)
	}

	// load states
	var blockEntities map[*big.Int]database.BlockEntity
	blockRows, err := database.DB.Queryx("SELECT * FROM blocks ORDER BY block_number")
	if err != nil {
		return err
	}

	for blockRows.Next() {
		// insert blocks
		var block database.BlockEntity
		if err := blockRows.StructScan(&block); err != nil {
			log.Errorf("failed to scan struct of userId %s: %v ", *block.UserId, err)
			return err
		}
		blockNumber := big.NewInt(*block.Number)
		blockEntities[blockNumber] = block
	}

	for blockNumber, block := range blockEntities {
		chain := sm.GetChain(*block.UserId)
		if _, ok := chain.blocks[blockNumber]; ok {
			// no need to update
			continue
		}

		newState := NewState()
		if err := newState.FromBytes(block.State); err != nil {
			log.Errorf("failed to parse state of userId %s: %v", *block.UserId, err)
			return err
		}

		// find transaction
		tx, exists := transactions[*block.TxHash]
		if !exists {
			log.Errorf("transaction not found for tx_hash %s: %v", *block.TxHash, err)
			return err
		}

		prevBlockHash, err := hexToHash(*block.PrevBlockHash)
		if err != nil {
			log.Errorf("failed to parse prev_block_hash of userId %s: %v", *block.UserId, err)
			return err
		}
		
		newBlock := NewBlock(blockNumber, newState, tx, prevBlockHash)
		chain.InsertBlock(newBlock)
	}

	return nil
}
