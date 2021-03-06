# 🎓 Quick Start
```go
package main

import (
	"github.com/HOWZ1T/space_trader"
	"fmt"
)

func check(err error) {
	if err != nil {
		panic(err)
    }
}

func main() {
    spaceTrader := space_trader.New("<token>", "<username>")
    status, err := spaceTrader.ApiStatus()
    check(err)
    fmt.Printf("API Status: %s", status)
}
```
