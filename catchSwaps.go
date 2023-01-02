package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const Uniswap = "0x0d4a11d5EEaaC28EC3F61d100daF4d40471f1852"
const Sushiswap = "0x06da0fd433C1A5d7a4faa01111c044910A184553"

func pack_swap(swapEvent handSwapEvent, blockNumber uint64, pair_name string, flag bool) (uint64, swap_data) {
	In := new(big.Float)
	Out := new(big.Float)
	course := new(big.Float)
	Buy := swapEvent.Amount0In.Cmp(big.NewInt(0)) == 0
	if Buy {
		In.SetInt(swapEvent.Amount1In)
		Out.SetInt(swapEvent.Amount0Out)
		trade = "Buy"
	} else {
		In.SetInt(swapEvent.Amount0In)
		Out.SetInt(swapEvent.Amount1Out)
		trade = "Sell"
	}

	if (Buy && !flag) || (!Buy && flag) {
		In.Quo(In, eth_decimals)
		Out.Quo(Out, usd_decimals)
		course = course.Quo(Out, In)
	} else {
		In.Quo(In, usd_decimals)
		Out.Quo(Out, eth_decimals)
		course = course.Quo(In, Out)
	}

	swap := swap_data{
		trade:     trade,
		course:    course.String(),
		in:        In.String(),
		out:       Out.String(),
		pair_name: pair_name,
	}
	return blockNumber, swap
}

func getSwaps(client *ethclient.Client, block uint64, address string, pair_name string, flag bool) map[uint64][]swap_data {
	contractAddress := common.HexToAddress(address)
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(block - 1000),
		ToBlock:   new(big.Int).SetUint64(block),
		Addresses: []common.Address{
			contractAddress,
		},
	}
	swaps_data := map[uint64][]swap_data{}
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	logSwapSig := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
	logSwapSigHash := crypto.Keccak256Hash(logSwapSig)
	for _, vLog := range logs {
		if vLog.Topics[0].Hex() == logSwapSigHash.Hex() {
			In0 := new(big.Int)
			In1 := new(big.Int)
			Out0 := new(big.Int)
			Out1 := new(big.Int)
			In0.SetBytes(vLog.Data[0:32])
			In1.SetBytes(vLog.Data[32:64])
			Out0.SetBytes(vLog.Data[64:96])
			Out1.SetBytes(vLog.Data[96:128])

			handSwap := handSwapEvent{In0, In1, Out0, Out1}
			blockNumber, swap := pack_swap(handSwap, vLog.BlockNumber, pair_name, flag)
			swaps_data[blockNumber] = append(swaps_data[blockNumber], swap)
		}
	}
	return swaps_data
}

func getData(client *ethclient.Client) (map[uint64][]swap_data, map[uint64][]swap_data) {
	block, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("error to get a block:%v", err)
	}
	uniswaps := getSwaps(client, block, Uniswap, "Uniswap", true)
	sushiswaps := getSwaps(client, block, Sushiswap, "Sushiswap", true)
	return uniswaps, sushiswaps
}
