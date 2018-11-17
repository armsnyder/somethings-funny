package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("Hello world")
		time.Sleep(2 * time.Second)
	}
}
