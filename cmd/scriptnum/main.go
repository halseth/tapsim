package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/halseth/mattlab/commitment"
)

func main() {
	num, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	s := commitment.ScriptNum(num)

	fmt.Printf("%x\n", s.Bytes())

}
