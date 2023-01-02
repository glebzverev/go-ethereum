package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var (
	ALCHEMY_KEY string
	client      *ethclient.Client
)

func dotenv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func init() {

	eth_decimals = big.NewFloat(math.Pow10(18))
	usd_decimals = big.NewFloat(math.Pow10(6))

	dotenv()
	ALCHEMY_KEY = os.Getenv("ALCHEMY_KEY")
	var err error
	client, err = ethclient.Dial(ALCHEMY_KEY)
	if err != nil {
		log.Fatal(err)
	}
}

func response(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request from:\t", r.RemoteAddr)
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

func main() {
	uniswaps, sushiswaps := getData(client)
	writer(uniswaps, "uniswap")
	writer(sushiswaps, "sushiswap")

	http.HandleFunc("/giveLast1000", response)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(":3000", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
