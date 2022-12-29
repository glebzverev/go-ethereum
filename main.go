package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func dotenv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	dotenv()
	ALCHEMY_KEY := os.Getenv("ALCHEMY_KEY")

	client, err := ethclient.DialContext(context.Background(), ALCHEMY_KEY)
	if err != nil {
		log.Fatalf("Error to create a ethere client:%v", err)
	}
	defer client.Close()
	block, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("error to get a block:%v", err)
	}
	fmt.Println(block)
}
