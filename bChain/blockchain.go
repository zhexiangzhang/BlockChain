package bChain

import (
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"os"
	"runtime"
)

type BlockChain struct {
	LastHash []byte     // 最后一个block的hash，存在数据库中
	Database *badger.DB // 指向数据库的指针
}

// 用于遍历区块链
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST" // 用于判断数据库是否存在
	genesisData = "First Block in the Chain - Genesis"
)

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte // 最后一个block的hash，内存中的变量

	// 判断数据库是否存在
	if DBexists() {
		fmt.Println("Blockchain already exists.")
		runtime.Goexit()
	}

	// NO blockchain in the database

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath      // store key and metadata in this directory
	opts.ValueDir = dbPath // store values in this directory

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		// create a coinbase transaction
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err := txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash
		return err
	})

	Handle(err)
	// create blockchain in memory
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func ContinueBlockChain(address string) *BlockChain {
	if DBexists() == false {
		fmt.Println("No existing blockchain found. Create one first!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath      // store key and metadata in this directory
	opts.ValueDir = dbPath // store values in this directory

	db, err := badger.Open(opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		// get last hash from database
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})

	Handle(err)
	// create blockchain in memory
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(transaction []*Transaction) {
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

	newBlock := CreateBlock(transaction, lastHash)

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

// Unspent means that these outputs weren’t referenced in any input

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction

	// 存储已花费的交易输出，也就是transaction input地址为address的交易
	// key是input的交易ID，value是input的交易输出索引
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()
		// 遍历区块中的所有交易
		for _, tx := range block.Transaction {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			// 注意：区块链遍历是从后往前。最后面的output肯定没有被花费，而最后交易的input可能是前面交易的output。
			// 但我们一定会先查看到input(已经被花费) ，然后记录到spentTXOs中
			for outIdx, out := range tx.Outputs {
				// Was the output spent?
				// 如果spentTXOs中有交易ID，就遍历这个交易ID对应的所有输出索引
				// 这个if的目的是如果这个交易花出去了就不会再记录到unspentTXs中
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				// 如果输出的地址和传入的地址相同，就把这个交易添加到unspentTXs中
				// 也就是说，out指定的输出对象pubkey就是这个接受者address，name这个交易就算address的收入，将其加入其unspentTXs中
				if out.CanBeUnlocked(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				// 遍历所有的输入，如果输入的地址和传入的地址相同，就把输入的交易ID和输出的索引添加到spentTXOs中
				// 输入说明，这是这个address花的钱
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						// in.Out记录了输出给了哪些
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// The function FindUnspentTransactions returns a list of transactions containing unspent outputs.
// To calculate balance we need takes the transactions and returns only outputs:
func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// for normal transaction, not coinbase
/*
The method iterates over all unspent transactions and accumulates their values.
When the accumulated value is more or equals to the amount we want to transfer,
it stops and returns the accumulated value and output indices.

*/
func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := chain.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
