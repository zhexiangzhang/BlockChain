package bChain

import (
	"bytes"
	"encoding/gob"
	"log"
)

//type BlockChain struct {
//	Blocks []*Block // which contais one field which has an array of pointers to blocks
//}

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

func CreateBlock(data string, preHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), preHash, 0} // 怎么转换
	//block.DeriveHash()
	// run the proof if work algorithm on each block we create
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

//func (chain *BlockChain) AddBlock(data string) {
//	preBlock := chain.Blocks[len(chain.Blocks)-1]
//	new := CreateBlock(data, preBlock.Hash)
//	chain.Blocks = append(chain.Blocks, new)
//}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

//func InitBlockChain() *BlockChain {
//	// 先生成 *Block
//	genesisBlock := Genesis()
//	// 再生成 []*Block
//	blockLink := []*Block{genesisBlock}
//	// 最后生成需要的blockchain类型
//	return &BlockChain{blockLink} // blockchain类型: blocks []*Block
//}
