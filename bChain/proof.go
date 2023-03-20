package bChain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

// proof of work

// Take the data from the block

// create a counter(nonce) which starts at 0

// creare a hash of the data plus the counter

// check the hash to see if it meets a set of requirments
// The frist few bytes must contain 0s

const Difficulty = 18

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// 满足条件的hash空间 T
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty)) //左移  256是hash的总位数
	// target是恰巧是最小的可能满足要求的hash，前12为都为0，剩下的256-12位可以从10000000开始便利
	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte
	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println()

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	// 把nonce传入 拿到header的整体组合
	data := pow.InitData(pow.Block.Nonce)

	// 验证使用这个nonce后header的hash是否满足条件
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

// 接受一个int64类型的参数num, 返回num的大端序十六进制表示形式。
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer) // 创建一个buffer缓冲区对象，该对象可以方便地进行二进制数据的读写
	// 将num按照大端序的格式写入到buff中。如果在写入数据的过程中出现了错误，err会非空
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	} //将buff对象中存储的字节切片返回
	return buff.Bytes()
}
