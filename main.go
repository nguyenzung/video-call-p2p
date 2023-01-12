package main

import (
	"fmt"

	"github.com/nguyenzung/nodego/runtimeutils"
)

func main() {
	fmt.Println("VideoCall server", runtimeutils.ThreadID())
	service := MakeService()
	service.Exec()
}
