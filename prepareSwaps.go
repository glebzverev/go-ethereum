package main

import (
	"math/big"
)

var (
	eth_decimals *big.Float
	usd_decimals *big.Float
	trade        string
)

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
			same_side_swaps[keys[i]] = append(same_side_swaps[keys[i]], "Buy")
		}
		if haveSide("Sell", uni[keys[i]]) && haveSide("Sell", sushi[keys[i]]) {
			same_side_swaps[keys[i]] = append(same_side_swaps[keys[i]], "Sell")
		}
	}
	return same_side_swaps
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
