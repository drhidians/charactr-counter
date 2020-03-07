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
	"sync"
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

	var runes = make(map[rune]uint)

	var wg sync.WaitGroup
	var m sync.Mutex

	err := filepath.Walk(*root,

		func(path string, info os.FileInfo, err error) error {

			if err != nil || info.IsDir() {
				return err
			}

			wg.Add(1)
			go func(file string) {

				defer wg.Done()

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
						m.Lock()
						runes[c]++
						m.Unlock()
					}
				}
			}(path)

			return nil
		})

	if err != nil {
		usage()
		log.Println(err)
		os.Exit(1)
	}

	wg.Wait()

	for k, v := range runes {
		fmt.Printf("%q: %d\n", k, v)
	}

}
