package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"math"
	"os"
	"sync"
)

func main() {
	flag.Parse()
	file := flag.Args()[0]

	info, err := os.Stat(file)
	if err != nil {
		panic(err)
	}

	fileSize := info.Size()
	workerCount := 8
	seekBase := fileSize / int64(workerCount)

	type readTask struct {
		file      *os.File
		buffer    []byte
		bytesLeft int64
	}
	tasks := make(chan *readTask, workerCount)

	var nextStart int64 = 0
	for i := range int64(workerCount) {
		file, err := os.OpenFile(file, os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}

		buffer := make([]byte, 512*1024)

		// Read last file til the end.
		if i == int64(workerCount)-1 {
			if _, err := file.Seek(nextStart, 0); err != nil {
				panic(err)
			}
			tasks <- &readTask{file, buffer, math.MaxInt64}
			continue
		}

		// maxlen Name;NN.N\n (106 bytes at max per line)
		end := nextStart + (seekBase - 106)
		if _, err := file.Seek(end, 0); err != nil {
			panic(err)
		}

		if _, err := file.Read(buffer); err != nil {
			panic(err)
		}

		for index, b := range buffer {
			if b == '\n' {
				end += int64(index) + 1
				break
			}
		}

		if _, err := file.Seek(nextStart, 0); err != nil {
			panic(err)
		}

		tasks <- &readTask{file, buffer, end - nextStart + 1}
		nextStart = end
	}
	close(tasks)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for task := range tasks {
		go func() {
			defer wg.Done()

			var last []byte

			type station struct {
				min, max int16
				sum      int16
				count    int64
			}
			// FIXME Optimise parsing into int16 instead of float.
			m := make(map[string]*station, 10000)

			for {
				if task.bytesLeft <= 0 {
					break
				}

				bytesRead, err := task.file.Read(task.buffer)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					panic(err)
				}

				if bytesRead >= int(task.bytesLeft) {
					bytesRead = int(task.bytesLeft)
					task.bytesLeft = 0
				} else {
					task.bytesLeft -= int64(bytesRead)
				}

				// reslice to treat last line not being read fully.
				buffer := append(last, task.buffer[:bytesRead]...)

				from := 0
				for i := 0; i < len(buffer); i++ {
					if buffer[i] == '\n' {
						// Drop newline for new slice
						line := buffer[from:i]
						// fmt.Println(string(line))

						semicolonIndex := bytes.IndexByte(line, ';')
						name := line[:semicolonIndex]
						number := line[semicolonIndex+1:]

						value := parseNumber(number)
						val := m[string(name)]
						if val == nil {
							val = &station{}
							m[string(name)] = val
						}

						if value < val.min {
							val.min = value
						}
						if value > val.max {
							val.max = value
						}
						val.sum += value
						val.count++

						from = i + 1
					}
				}
				last = buffer[from:]
			}
		}()
	}

	wg.Wait()
}

func parseNumber(bytes []byte) int16 {
	weight := len(bytes) - 2
	var value int16
	for _, b := range bytes {
		if b == '-' {
			weight = weight - 1
			continue
		} else if b == '.' {
			continue
		}

		digit := b - 48
		for range weight {
			digit = digit * 10
		}
		weight = weight - 1
		value += int16(digit)
	}

	if bytes[0] == '-' {
		return -1 * value
	}

	return value
}
