package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/halseth/mattlab/commitment"
)

func main() {
	val := os.Args[1]
	var (
		num uint64
		err error
	)

	if val == "from" {
		scriptNumHex := os.Args[2]
		v, err := hex.DecodeString(scriptNumHex)
		if err != nil {
			panic(err)
		}

		scriptNum, err := commitment.MakeScriptNum(v, true, len(v))
		if err != nil {
			panic(err)
		}

		fmt.Printf("%d\n", scriptNum)
		return
	}

	base := 10
	if strings.HasPrefix(val, "0x") {
		base = 16
		val = strings.TrimPrefix(val, "0x")
	}
	num, err = strconv.ParseUint(val, base, 64)
	if err != nil {
		panic(err)
	}
	s := commitment.ScriptNum(num)

	fmt.Printf("%x\n", s.Bytes())

}
