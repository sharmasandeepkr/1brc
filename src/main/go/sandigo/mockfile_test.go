package main

import (
	"fmt"
	"os"
	"testing"
)

func TestMockFile(t *testing.T) {
	var file mockFile
	var err error
	file, err = os.Open("./s.text")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := processFile(file)
	if err != nil {
		panic(err)
	}
	fmt.Println("read this much byte ", data)
}
