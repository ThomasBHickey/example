package main

import (
	"fmt"
	"github.com/ThomasBHickey/metatrue"
)

func main() {
	fmt.Println("mtTry calling metatrue.Start()")
	err := metatrue.Start()
	if err != nil {
		fmt.Println("called Start OK", err)
	} else {
		fmt.Println("Start returned OK")
	}
}
