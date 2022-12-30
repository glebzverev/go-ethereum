package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"main/pair"
	"math"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

const Uniswap = "0x0d4a11d5EEaaC28EC3F61d100daF4d40471f1852"
const Sushiswap = "0x06da0fd433C1A5d7a4faa01111c044910A184553"
const eth_decimals = 18
const usd_decimals = 6

type swap_data struct {
	trade     string
	course    string
	in        string
	out       string
	pair_name string
}

func pack_swap(swapEvent pair.PairSwap, blockNumber uint64, pair_name string, flag bool) (uint64, swap_data) {
	if swapEvent.Amount0In.String() == "0" {
		In := new(big.Float)
		In.SetString(swapEvent.Amount1In.String())
		Out := new(big.Float)
		Out.SetString(swapEvent.Amount0Out.String())
		course := new(big.Float)
		if !flag {
			In = new(big.Float).Quo(In, big.NewFloat(math.Pow10(eth_decimals)))
			Out = new(big.Float).Quo(Out, big.NewFloat(math.Pow10(usd_decimals)))
			course = course.Quo(Out, In)
		} else {
			In = new(big.Float).Quo(In, big.NewFloat(math.Pow10(usd_decimals)))
			Out = new(big.Float).Quo(Out, big.NewFloat(math.Pow10(eth_decimals)))
			course = course.Quo(In, Out)
		}
		swap := swap_data{
			trade:     "Buy",
			course:    course.String(),
			in:        In.String(),
			out:       Out.String(),
			pair_name: pair_name,
		}
		return blockNumber, swap
	} else {
		In := new(big.Float)
		In.SetString(swapEvent.Amount0In.String())
		Out := new(big.Float)
		Out.SetString(swapEvent.Amount1Out.String())
		course := new(big.Float)
		if flag {
			In = new(big.Float).Quo(In, big.NewFloat(math.Pow10(eth_decimals)))
			Out = new(big.Float).Quo(Out, big.NewFloat(math.Pow10(usd_decimals)))
			course = course.Quo(Out, In)
		} else {
			In = new(big.Float).Quo(In, big.NewFloat(math.Pow10(usd_decimals)))
			Out = new(big.Float).Quo(Out, big.NewFloat(math.Pow10(eth_decimals)))
			course = course.Quo(In, Out)
		}
		swap := swap_data{
			trade:     "Sell",
			course:    course.String(),
			in:        In.String(),
			out:       Out.String(),
			pair_name: pair_name,
		}
		return blockNumber, swap
	}
}

func dotenv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
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

	contractAbi, err := abi.JSON(strings.NewReader(string(pair.PairABI)))
	if err != nil {
		log.Fatal(err)
	}

	logSwapSig := []byte("Swap(address,uint256,uint256,uint256,uint256,address)")
	logSwapSigHash := crypto.Keccak256Hash(logSwapSig)
	for _, vLog := range logs {
		switch vLog.Topics[0].Hex() {
		case logSwapSigHash.Hex():
			var swapEvent pair.PairSwap
			err := contractAbi.UnpackIntoInterface(&swapEvent, "Swap", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			blockNumber, swap := pack_swap(swapEvent, vLog.BlockNumber, pair_name, flag)
			swaps_data[blockNumber] = append(swaps_data[blockNumber], swap)
		}
	}
	return swaps_data
}

func response(w http.ResponseWriter, r *http.Request) {
	dotenv()
	ALCHEMY_KEY := os.Getenv("ALCHEMY_KEY")
	client, err := ethclient.Dial(ALCHEMY_KEY)
	if err != nil {
		log.Fatal(err)
	}
	uniswaps, sushiswaps := getData(client)

	unikeys := make([]uint64, 0, len(uniswaps))
	for k := range uniswaps {
		unikeys = append(unikeys, k)
	}
	sushikeys := make([]uint64, 0, len(sushiswaps))
	for k := range sushiswaps {
		sushikeys = append(sushikeys, k)
	}
	concatkeys := append(unikeys, sushikeys...)
	sort.Slice(concatkeys, func(i, j int) bool { return concatkeys[i] < concatkeys[j] })
	concatkeys = dropDuplicate(concatkeys)
	same_side_swaps := sameSide(concatkeys, uniswaps, sushiswaps)

	fmt.Fprintf(w, "%++v", uniswaps)
	fmt.Fprintf(w, "%++v", sushiswaps)
	fmt.Fprintf(w, "%++v", same_side_swaps)
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

func dropDuplicate(s []uint64) []uint64 {
	if len(s) < 2 {
		return s
	}
	tmp := []uint64{}
	for i := 0; i < len(s)-1; i++ {
		if s[i] != s[i+1] {
			tmp = append(tmp, s[i])
		}
	}
	if s[len(s)-2] != s[len(s)-1] {
		tmp = append(tmp, s[len(s)-1])
	}
	return tmp
}

func haveSide(side string, swaps []swap_data) bool {
	for i := range swaps {
		if swaps[i].trade == side {
			return true
		}
	}
	return false
}

func sameSide(keys []uint64, uni map[uint64][]swap_data, sushi map[uint64][]swap_data) map[uint64][]string {
	same_side_swaps := map[uint64][]string{}
	for i := range keys {
		if haveSide("Buy", uni[keys[i]]) && haveSide("Buy", sushi[keys[i]]) {
			fmt.Println(keys[i], "Buy")
			same_side_swaps[keys[i]] = append(same_side_swaps[keys[i]], "Buy")
		}
		if haveSide("Sell", uni[keys[i]]) && haveSide("Sell", sushi[keys[i]]) {
			fmt.Println(keys[i], "Sell")
			same_side_swaps[keys[i]] = append(same_side_swaps[keys[i]], "Sell")
		}
	}
	return same_side_swaps
}

func main() {
	http.HandleFunc("/giveLast1000", response)
	err := http.ListenAndServe(":80", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
