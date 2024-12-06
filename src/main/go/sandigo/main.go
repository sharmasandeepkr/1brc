package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"
)

type stationData struct {
	min   float64
	max   float64
	sum   float64
	count int64
}

type store struct {
	locker    *sync.Mutex
	globalMap map[string]*stationData
}

const measurementsPath string = "./new1b.text"

var (
	// others: "heap", "threadcreate", "block", "mutex"
	shouldProfile bool = false
	profileTypes       = []string{"goroutine", "allocs"}
)

func main() {
	if shouldProfile {
		nowUnix := time.Now().Unix()
		os.MkdirAll(fmt.Sprintf("profiles/%d", nowUnix), 0755)
		for _, profileType := range profileTypes {
			file, _ := os.Create(fmt.Sprintf("profiles/%d/%s.%s.pprof",
				nowUnix, filepath.Base(measurementsPath), profileType))
			defer file.Close()
			defer pprof.Lookup(profileType).WriteTo(file, 0)
		}

		file, _ := os.Create(fmt.Sprintf("profiles/%d/%s.cpu.pprof",
			nowUnix, filepath.Base(measurementsPath)))
		defer file.Close()
		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}

	start := time.Now()
	var newStore store = store{
		locker:    &sync.Mutex{},
		globalMap: make(map[string]*stationData),
	}
	// job := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		// go Caps  uleWorker(ctx, job, &newStore, wg)
	}
	file, err := os.Open(measurementsPath)
	if err != nil {
		ctx.Done()
		panic(err)
	}
	defer file.Close()
	// ReadFile(ctx, file, job)
	wg.Wait()
	for key, stationData := range newStore.globalMap {
		fmt.Println("x:", key, stationData)
	}
	// keyByte := []byte("Hamburg")
	// hash := fnv.New64a() // 64-bit FNV-1a hash
	// hash.Write(keyByte)
	// key := hash.Sum64()
	// v, ok := newStore.globalMap[key]
	// if !ok {
	// 	panic(0)
	// }
	// fmt.Println(v)
	fmt.Println(time.Since(start))
}

func MapMerger(oldStore *store, incomming map[string]*stationData) {
	oldStore.locker.Lock()
	defer oldStore.locker.Unlock()
	for keyInc, valueInc := range incomming {
		storeValue, ok := oldStore.globalMap[keyInc]
		if !ok {
			oldStore.globalMap[keyInc] = valueInc
			continue
		}
		storeValue.count += valueInc.count

		if storeValue.max < valueInc.max {
			storeValue.max = valueInc.max
		}
		if storeValue.min > valueInc.min {
			storeValue.min = valueInc.min
		}
		storeValue.sum += valueInc.sum
	}
}

// parseFloatFast is a high performance float parser using the assumption that
// the byte slice will always have a single decimal digit.
func parseFloatFast(bs []byte) float64 {
	var intStartIdx int // is negative?
	if bs[0] == '-' {
		intStartIdx = 1
	}

	v := float64(bs[len(bs)-1]-'0') / 10 // single decimal digit
	place := 1.0
	for i := len(bs) - 3; i >= intStartIdx; i-- { // integer part
		v += float64(bs[i]-'0') * place
		place *= 10
	}

	if intStartIdx == 1 {
		v *= -1
	}
	return v
}

func insertKeys(keyByte []byte, temperatureByte []byte, tempMap map[string]*stationData) {
	key := string(keyByte)
	temperature := parseFloatFast(temperatureByte)

	sd, ok := tempMap[key]
	if !ok {
		tempMap[key] = &stationData{
			min:   temperature,
			max:   temperature,
			sum:   temperature,
			count: 1,
		}
		return
	}
	sd.count += 1
	sd.sum += temperature
	if temperature < sd.min {
		sd.min = temperature
		return
	}
	if temperature > sd.max {
		sd.max = temperature
	}
}

func removeIf5Zero(buf []byte, idx int64) []byte {
	count := 0
	for idx < int64(len(buf)) {
		if count == 5 {
			return buf[:idx-5]
		}
		if buf[idx] == 0 {
			count++
		}
		if buf[idx] != 0 && count > 0 {
			count = 0
		}
		idx++
	}
	return buf
}

func ParseAt(f *os.File, offset int64, chunkSize int64, locker *sync.Mutex) {
	buf := make([]byte, chunkSize+120)
	var tempMap = make(map[string]*stationData)
	// defer func() {
	// 	locker.Lock()
	// 	fmt.Println(string(buf))
	// 	for key, value := range tempMap {
	// 		fmt.Println("x: ", key, value)
	// 	}
	// 	fmt.Println("----------", offset)
	// 	locker.Unlock()
	// }()
	_, err := f.ReadAt(buf, offset)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	var idx, si, nli, nliNext int64
	var exist = false
	for idx < int64(len(buf)) {
		if offset == 0 && idx == 0 {
			nli = 0
			//no need to find \n to neglect bytes
			// find ; after index idx which return si if exist
			si, exist = find(buf, idx, ';')
			if !exist {
				fmt.Println("could not find ; when offset =0")
				return
			}
			idx = si
			nli, exist = find(buf, idx, '\n')
			if !exist {
				insertKeys(buf[0:si], removeIf5Zero(buf[si+1:], 0), tempMap)
				return
			}
			idx = nli
			insertKeys(buf[0:si], buf[si+1:nli], tempMap)
			si = 0
		}

		// chop  first \n in case it is not offset = 0
		if offset != 0 && idx == 0 {
			nli, exist = find(buf, idx, '\n')
			if !exist {
				return
			}
			idx = nli
		}
		si, exist = find(buf, idx, ';')
		if !exist {
			fmt.Println("when offset is not 0, and idx != 0 but not able to find ;")
			return
		}
		idx = si
		nliNext, exist = find(buf, idx, '\n')
		if !exist {
			//check fot the last line irrespectice of ofset
			insertKeys(buf[nli+1:si], removeIf5Zero(buf[si+1:], 0), tempMap)
			return
		}

		if nli > si {
			fmt.Println("why??", nli, si, offset, string(buf))
		}
		// fmt.Println(nli+1, si, si+1, nliNext)
		insertKeys(buf[nli+1:si], buf[si+1:nliNext], tempMap)
		if idx > chunkSize {
			return
		}
		idx = nliNext
		nli = nliNext
		nliNext = 0
		si = 0
		idx++
	}

}

func find(buf []byte, lidx int64, letter rune) (int64, bool) {
	for lidx < int64(len(buf)) {
		if buf[lidx] == byte(letter) {
			return lidx, true
		}
		lidx++
	}
	return 0, false
}

// func LineWriterWorker(ctx context.Context, lineWriter chan []byte, singleStore *store, wg *sync.WaitGroup) {
// 	tempMap := make(map[string]*stationData)
// 	defer wg.Done()
// 	ticker := time.NewTicker(500 * time.Millisecond)
// 	flush := func() {
// 		MapMerger(singleStore, tempMap)
// 		tempMap = make(map[string]*stationData)
// 	}
// 	defer flush()

// 	insertKeys := func(key string, temperature float64) {
// 		sd, ok := tempMap[key]
// 		if !ok {
// 			tempMap[key] = &stationData{
// 				min:   temperature,
// 				max:   temperature,
// 				sum:   temperature,
// 				count: 1,
// 			}
// 			return
// 		}
// 		sd.count += 1
// 		sd.sum += temperature
// 		if temperature < sd.min {
// 			sd.min = temperature
// 			return
// 		}
// 		if temperature > sd.max {
// 			sd.max = temperature
// 		}
// 	}

// 	for {
// 		select {
// 		case _, ok := <-ticker.C:
// 			{
// 				if !ok {
// 					return
// 				}
// 				flush()
// 			}
// 		case strCapsule, ok := <-lineWriter:
// 			{
// 				if !ok {
// 					return
// 				}
// 				// fmt.Println(string(line))

// 				// allLines := strings.Split(strCapsule, "\n")
// 				var lastIndex int = 0
// 				var key string
// 				for i := 0; lastIndex+i < len(strCapsule); i++ {
// 					if strCapsule[lastIndex+i] == ';' {
// 						key = string(strCapsule[lastIndex : lastIndex+i])
// 						var floatByte []byte
// 						var j int
// 						for j = 1; lastIndex+i+j < len(strCapsule) && strCapsule[lastIndex+i+j] != '\n'; j++ {
// 							floatByte = append(floatByte, strCapsule[lastIndex+i+j])
// 						}
// 						insertKeys(key, parseFloatFast(floatByte))
// 						i = 0
// 						lastIndex += i + j + 1
// 					}
// 				}
// 			}

// 		case <-ctx.Done():
// 			return
// 		}
// 	}
// }

// // worker keeps on working and store in a temp memory and in aysnc way, once write is availbel to it, it flushes its buffer
// func CapsuleWorker(ctx context.Context, job chan []byte, singleStore *store, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	lineWriter := make(chan []byte)
// 	defer close(lineWriter)

// 	for i := 0; i < 3; i++ {
// 		wg.Add(1)
// 		go LineWriterWorker(ctx, lineWriter, singleStore, wg)
// 	}

// 	for {
// 		select {
// 		case byteCapsule, ok := <-job:
// 			if !ok {
// 				return
// 			}
// 			// strCapsule := string(byteCapsule)
// 			lineWriter <- byteCapsule

// 		case <-ctx.Done():
// 			return

// 		}
// 	}
// }

// func ReadFile(ctx context.Context, file mockFile, job chan []byte) {
// 	newData := func() []byte {
// 		return make([]byte, 3*1024)
// 	}
// 	data := newData()
// 	lastn := 0
// 	defer close(job)
// 	EOF := false
// 	for {
// 		if EOF {
// 			return
// 		}
// 		n, err := file.ReadAt(data, int64(lastn))
// 		if err != nil {
// 			if err == io.EOF {
// 				EOF = true
// 			} else {
// 				break
// 			}
// 		}
// 		lastn += n
// 		newLineByte := []byte("\n")[0]
// 		if data[len(data)-1] != newLineByte {
// 			singleByte := make([]byte, 1)
// 			for singleByte[0] != newLineByte && !EOF {
// 				// if EOF {
// 				// 	break
// 				// }
// 				n, err := file.ReadAt(singleByte, int64(lastn))
// 				if err != nil {
// 					if err == io.EOF {
// 						EOF = true
// 					} else {
// 						break
// 					}
// 				}
// 				lastn += n
// 				data = append(data, singleByte...)
// 			}
// 		}

// 		select {
// 		case job <- data:
// 			data = newData()

// 		case <-ctx.Done():
// 			return
// 		}
// 	}
// }
