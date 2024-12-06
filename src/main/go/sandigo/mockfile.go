package main

import (
	"fmt"
	"io"
	"os"
)

type mockFile interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}

// type testFile struct {
// 	content []byte
// }

// func (t *testFile) Read(b []byte) (int, error) {
// 	fmt.Println("why i am being called")
// 	b = t.content
// 	return len(t.content), nil
// }

// func (t *testFile) Write(data []byte) (int, error) {
// 	t.content = data
// 	return len(data), nil
// }

// func (t *testFile) Close() error {
// 	return nil
// }

func processFile(file mockFile) ([]byte, error) {
	data := make([]byte, 1)
	totalData := []byte{}
	lastn := 0
	for {
		n, err := file.ReadAt(data, int64(lastn))
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		lastn += n
		totalData = append(totalData, data...)
	}
	if len(totalData) == 11 {
		fmt.Println("hurrey, going well")
	}
	return totalData, nil
}

// func main() {
// 	var file mockFile
// 	var err error
// 	file, err = os.Open("./s.text")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()
// 	data, err := processFile(file)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("read this much byte ", data)
// 	fmt.Println("read this much byte ", string(data))

// }
