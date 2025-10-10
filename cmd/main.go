package main

import (
	"fmt"

	"github.com/s-588/BOMViewer/cmd/http"
)

func main() {
	fmt.Println(http.NewServer(":8080").Start())
}
