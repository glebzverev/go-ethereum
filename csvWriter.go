package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func writer(records map[uint64][]swap_data, name string) {
	// Create a csv file
	f, err := os.Create("./static/" + name + ".csv")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	for num, obj := range records {
		for r := range obj {
			var record []string
			record = append(record, strconv.Itoa(int(num)))
			record = append(record, obj[r].course)
			record = append(record, obj[r].trade)
			record = append(record, obj[r].pair_name)
			w.Write(record)
		}
	}
	w.Flush()
}
