package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

// func TestReadFile(t *testing.T) {
// 	file, err := os.Open("./s.text")
// 	if err != nil {
// 		t.Fail()
// 	}
// 	defer file.Close()
// 	job := make(chan []byte)
// 	ctx := context.Background()
// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		for {
// 			select {
// 			case capsule, ok := <-job:
// 				{
// 					if !ok {
// 						return
// 					}
// 					fmt.Println(capsule)
// 				}
// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 	}()
// 	ReadFile(ctx, file, job)
// 	wg.Wait()

// }

// func TestWorker(t *testing.T) {
// 	file, err := os.Open("./s.text")
// 	if err != nil {
// 		t.Fail()
// 	}
// 	defer file.Close()
// 	job := make(chan []byte)
// 	ctx := context.Background()
// 	wg := &sync.WaitGroup{}

// 	globalStore := store{
// 		locker:    &sync.Mutex{},
// 		globalMap: make(map[string]*stationData),
// 	}
// 	wg.Add(1)
// 	go CapsuleWorker(ctx, job, &globalStore, wg)
// 	ReadFile(ctx, file, job)
// 	wg.Wait()
// 	fmt.Println("len of map is ", len(globalStore.globalMap))
// 	for key, stationData := range globalStore.globalMap {
// 		fmt.Println("x:", key, stationData)
// 	}
// 	// keyByte := []byte("Hamburg")
// 	// hash := fnv.New64a() // 64-bit FNV-1a hash
// 	// hash.Write(keyByte)
// 	// key := hash.Sum64()
// 	// v, ok := globalStore.globalMap[key]
// 	// if !ok {
// 	// 	t.Fail()
// 	// }
// 	// fmt.Println(v)
// }

func TestGen(t *testing.T) {
	file, err := os.Create("new1b.text")
	if err != nil {
		t.Fail()
	}
	defer file.Close()

	for i := 0; i < 1000000*50; i++ {
		file.Write([]byte("Roseau;34.4\n"))
	}
	for i := 0; i < 1000000*50; i++ {
		file.Write([]byte("Bridgetown;26.9\n"))
	}
}

// func TestLineWriteWorker(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	store := store{
// 		locker:    &sync.Mutex{},
// 		globalMap: make(map[string]*stationData),
// 	}
// 	lineWriter := make(chan []byte)

// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)
// 	go LineWriterWorker(ctx, lineWriter, &store, wg)

// 	for i := 0; i < 10; i++ {
// 		lineWriter <- []byte("Roseau;34.4")
// 	}
// 	close(lineWriter)
// 	wg.Wait()
// 	for key, value := range store.globalMap {
// 		fmt.Println(key, "xx", value)
// 	}
// }

func TestMapMerger(t *testing.T) {
	// oldStore *store, incomming map[string]stationData)
	oldStore := &store{
		locker:    &sync.Mutex{},
		globalMap: make(map[string]*stationData),
	}
	incomming := make(map[string]*stationData)
	incomming["Hamburg"] = &stationData{
		min:   23.06,
		max:   23.06,
		count: 1,
		sum:   23.06,
	}
	MapMerger(oldStore, incomming)
	MapMerger(oldStore, incomming)
	MapMerger(oldStore, incomming)
	MapMerger(oldStore, incomming)
	for key, value := range oldStore.globalMap {
		fmt.Println(key, value)
	}
}

func TestReadAt(t *testing.T) {
	file, err := os.Open("./new1b.text")
	if err != nil {
		t.Fail()
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fail()
	}
	var chunkSize int64 = 100 * 1024

	offsetChan := make(chan int64, (info.Size()/chunkSize)+1)
	var i int64 = 0
	go func() {
		for i < info.Size() {
			offsetChan <- i
			i = i + chunkSize
		}
		close(offsetChan)
	}()
	wg := &sync.WaitGroup{}
	l := &sync.Mutex{}
	wg.Add(8)
	for i := 0; i < 8; i++ {
		go func() {
			defer wg.Done()
			for offSet := range offsetChan {
				ParseAt(file, offSet, chunkSize, l)
			}

		}()

	}
	wg.Wait()
}

func TestRemoval5zero(t *testing.T) {
	buf := make([]byte, 100)
	buf[0] = '1'
	buf[1] = '1'
	buf[2] = '.'
	buf[3] = '0'
	fmt.Println(string(removeIf5Zero(buf, 0)))

}

// func ParseAt(f *os.File, buf []byte, offset int64, chunkSize int64, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	_, err := f.ReadAt(buf, offset)
// 	if err != nil {
// 		if err != io.EOF {
// 			panic(err)
// 		}
// 	}
// 	var i int64
// 	var key, value []byte
// 	for i < int64(len(buf)) {
// 		if buf[i] == '\n' {
// 			if len(value) != 0 {
// 				key = append(key, buf[i])
// 			}
// 		}
// 		if buf[i]==';'{

// 		}
// 	}

// }
