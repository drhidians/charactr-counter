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

func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)
	go func() { // HL
		// Close the paths channel after Walk returns.
		defer close(paths) // HL
		// No select needed for this send, since errc is buffered.
		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error { // HL
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			paths <- path // HL

			return nil
		})
	}()
	return paths, errc
}

// runeReader reads path names from paths and sends character of the corresponding
// files on c until either paths or done is closed.
func runeReader(done <-chan struct{}, paths <-chan string, c chan<- rune) {
	for path := range paths { // HLpaths

		data, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		b := bufio.NewReader(bytes.NewReader(data))
		_ = b

		for {
			if r, _, err := b.ReadRune(); err != nil {
				if err == io.EOF {
					break
				}
				if err != nil {
					panic(err)
				}
			} else {
				c <- r
			}
		}

	}
}

// CharacterHistogram counts characters appearance in files under "root" directory
func CharacterHistogram(root string) (map[rune]uint, error) {
	// characterHistogram closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{})
	defer close(done)

	paths, errc := walkFiles(done, root)

	var numCpu = runtime.NumCPU()

	// Create fixed number of channels for better throughput
	c := make(chan rune, 20*numCpu) // HLc
	var wg sync.WaitGroup

	var numWorker = numCpu - 2
	wg.Add(numWorker)

	go func() {
		for i := 0; i < numWorker; i++ {
			runeReader(done, paths, c) // HLc
			wg.Done()
		}
	}()

	go func() {
		wg.Wait()
		close(c) // HLc
	}()
	// End of pipeline. OMIT

	m := make(map[rune]uint)
	for r := range c {
		m[r]++
	}
	// Check whether the Walk failed.
	if err := <-errc; err != nil { // HLerrc
		return nil, err
	}

	return m, nil
}

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
	t := time.Now()
	runes, err := CharacterHistogram(*root)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(time.Since(t))
	//return
	for r, v := range runes {
		fmt.Printf("%q: %d\n", r, v)
	}

}
