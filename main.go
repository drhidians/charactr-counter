package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// walkFiles starts a goroutine to walk the directory tree at root and send the
// path of each regular file on the string channel.  It sends the result of the
// walk on the error channel.  If done is closed, walkFiles abandons its work.
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
			select {
			case paths <- path: // HL
			case <-done: // HL
				return errors.New("walk canceled")
			}
			return nil
		})
	}()
	return paths, errc
}

func readRune(data []byte) <-chan rune {
	out := make(chan rune)
	r := bufio.NewReader(bytes.NewReader(data))

	go func() {
		for {
			if c, _, err := r.ReadRune(); err != nil {
				if err == io.EOF {
					break
				} else {
					panic(err)
				}
			} else {
				out <- c
			}
		}
	}()
	return out

}

// characterCounter reads path names from paths and sends character of the corresponding
// files on c until either paths or done is closed.
func characterCounter(done <-chan struct{}, paths <-chan string, c chan<- rune) {
	for path := range paths { // HLpaths
		data, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		b := bufio.NewReader(bytes.NewReader(data))

		for {
			if r, _, err := b.ReadRune(); err != nil {
				if err == io.EOF {
					break
				} else {
					panic(err)
				}
			} else {
				select {
				case c <- r:
				case <-done:
					return
				}

			}
		}

	}
}

func characterHistogram(root string) (map[rune]uint, error) {
	// characterHistogram closes the done channel when it returns; it may do so before
	// receiving all the values from c and errc.
	done := make(chan struct{})
	defer close(done)

	paths, errc := walkFiles(done, root)

	// Start a fixed number of goroutines to read and count characters from files.
	c := make(chan rune) // HLc
	var wg sync.WaitGroup
	const numWorker = 20
	wg.Add(numWorker)

	for i := 0; i < numWorker; i++ {
		go func() {
			characterCounter(done, paths, c) // HLc
			wg.Done()
		}()
	}

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

func main() {

	runes, err := characterHistogram(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	for r, v := range runes {
		fmt.Printf("%q: %d\n", r, v)
	}

}
