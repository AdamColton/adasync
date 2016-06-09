package main

import (
	"fmt"
	"os"
)

func main() {
	e := os.Mkdir("/does/not/exist/", 0777)
	fmt.Println(e)
}
