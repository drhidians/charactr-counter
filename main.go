package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func usage() {
	log.Printf("Usage: rune-counter [-p path] [-a amount]\n")
	flag.PrintDefaults()
}

func main() {

	var root = flag.String("p", "files", "Path to file")
	var amount = flag.Int("a", 0, "Amount of mock files to create inside Path")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *amount != 0 {
		createRandomFiles(*root, *amount)
	}

	var t = time.Now()

	var runes = make(map[rune]uint)

	var numCPU = runtime.NumCPU()
	var channel = make(chan rune, 100)

	done := make(chan bool, 1)

	go func() {

		for r := range channel {
			runes[r]++
		}
		done <- true
	}()

	var wg sync.WaitGroup
	//var m sync.Mutex

	// Keep alive "rune counter" goroutine
	var numWork = make(chan struct{}, numCPU)

	err := filepath.Walk(*root,

		func(path string, info os.FileInfo, err error) error {

			if err != nil || info.IsDir() {
				return err
			}

			wg.Add(1)

			numWork <- struct{}{}
			go func(file string) {

				defer func() {
					<-numWork
					defer wg.Done()
				}()

				content, err := ioutil.ReadFile(file)
				if err != nil {
					panic(err)
				}

				r := bufio.NewReader(bytes.NewReader(content))

				for {
					if c, _, err := r.ReadRune(); err != nil {
						if err == io.EOF {
							break
						} else {
							panic(err)
						}
					} else {
						channel <- c
					}
				}
			}(path)

			return nil
		})

	if err != nil {
		usage()
		os.Exit(1)
	}

	wg.Wait()
	close(channel)

	<-done
	close(done)

	fmt.Println(time.Since(t))
	return
	for k, v := range runes {
		fmt.Printf("%q: %d\n", k, v)
	}

}
