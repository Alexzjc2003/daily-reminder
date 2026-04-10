package main

import (
	"fmt"
	"time"
)

func main() {
	if time, err := time.Parse("2006/01/02", "0000/01/02"); err != nil {
		panic(err)
	} else {
		fmt.Printf("%v\n", time)
	}
}
