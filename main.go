package main

import (
	"flag"
	"fmt"
	"github.com/blockchain/bChain"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - Get the balance of an address")
	fmt.Println(" createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println(" print - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {
	chain := bChain.ContinueBlockChain("")
	defer chain.Database.Close()

	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := bChain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// address: the person who mine the genesis block
func (cli *CommandLine) createBlockChain(address string) {
	chain := bChain.InitBlockChain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) {
	chain := bChain.ContinueBlockChain(address)
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := bChain.ContinueBlockChain(from)
	defer chain.Database.Close()

	tx := bChain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*bChain.Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {

	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func main() {
	defer os.Exit(0)
	cli := CommandLine{}
	cli.run()
}

//type CommandLine struct {
//	blockchain *bChain.BlockChain
//}
//
//func (cli *CommandLine) printUsage() {
//	fmt.Println("Usage:")
//	fmt.Println(" addBlock -data BLOCK_DATA - add a block to the blockchain")
//	fmt.Println(" printChain - print all the blocks of the blockchain")
//}
//
//func (cli *CommandLine) validateArgs() {
//	if len(os.Args) < 2 {
//		cli.printUsage()
//		runtime.Goexit()
//	}
//}
//
//func (cli *CommandLine) addBlock(data string) {
//	cli.blockchain.AddBlock(data)
//	fmt.Println("Add Block!")
//}
//
//func (cli *CommandLine) printChain() {
//	iter := cli.blockchain.Iterator()
//
//	for {
//		block := iter.Next()
//
//		fmt.Printf("PrevHash: %x\n", block.PrevHash)
//		fmt.Printf("Data: %s\n", block.Data)
//		fmt.Printf("Hash: %x\n", block.Hash)
//
//		pow := bChain.NewProof(block)
//		fmt.Printf("Pow: %s \n", strconv.FormatBool(pow.Validate()))
//		fmt.Println()
//
//		if len(block.PrevHash) == 0 {
//			break
//		}
//	}
//}
//
//func (cli *CommandLine) run() {
//	cli.validateArgs()
//
//	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
//	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
//
//	addBlockData := addBlockCmd.String("block", "", "Block data")
//
//	switch os.Args[1] {
//	case "addBlock":
//		err := addBlockCmd.Parse(os.Args[2:])
//		bChain.Handle(err)
//	case "printChain":
//		err := printChainCmd.Parse(os.Args[2:])
//		bChain.Handle(err)
//	default:
//		cli.printUsage()
//		runtime.Goexit()
//	}
//
//	if addBlockCmd.Parsed() {
//		if *addBlockData == "" {
//			addBlockCmd.Usage()
//			runtime.Goexit()
//		}
//		cli.addBlock(*addBlockData)
//	}
//
//	if printChainCmd.Parsed() {
//		cli.printChain()
//	}
//}
//
//func main() {
//	//defer os.Exit(0)
//	chain := bChain.InitBlockChain()
//	defer chain.Database.Close()
//
//	cli := CommandLine{chain}
//	cli.run()
//}
