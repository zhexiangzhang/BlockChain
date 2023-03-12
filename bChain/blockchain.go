package bChain

import (
	"fmt"
	"github.com/dgraph-io/badger"
)

type BlockChain struct {
	LastHash []byte     // 最后一个block的hash，存在数据库中
	Database *badger.DB // 指向数据库的指针
	//Blocks   []*Block   // which contais one field which has an array of pointers to blocks
}

// 用于遍历区块链
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

const (
	dbPath = "./tmp/blocks"
)

func InitBlockChain() *BlockChain {
	var lastHash []byte // 最后一个block的hash，内存中的变量

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath      // store key and metadata in this directory
	opts.ValueDir = dbPath // store values in this directory

	db, err := badger.Open(opts)
	Handle(err)

	// whether there is a blockchain in the database
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			// no blockchain in the database
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")

			err := txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash
			return err
		} else {
			// there is a blockchain in the database
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			//err = item.Value(func(val []byte) error {
			//	lastHash = val
			//	return nil
			//})
			lastHash, err = item.ValueCopy(nil)
			return err
		}
	})

	Handle(err)
	// create blockchain in memory
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	//preBlock := chain.Blocks[len(chain.Blocks)-1]
	//new := CreateBlock(data, preBlock.Hash)
	//chain.Blocks = append(chain.Blocks, new)
	var lastHash []byte
	// view: read-only transaction
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		//err = item.Value(func(val []byte) error {
		//	lastHash = val
		//	return nil
		//})
		lastHash, err = item.ValueCopy(nil)
		return err
	})
	Handle(err)

	newBlock := CreateBlock(data, lastHash)

	// update: read-write transaction
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		//err = item.Value(func(val []byte) error {
		//	block = Deserialize(val)
		//	return nil
		//})
		encodedBlock, err := item.ValueCopy(nil)
		block = Deserialize(encodedBlock)

		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash
	return block
}
