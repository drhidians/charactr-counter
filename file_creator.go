package main

import (
	"math/rand"
	"os"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		_ = rand.Intn(128)
		b[i] = rune(67)
	}

	return string(b)
}

func createRandomFiles(root string, count int) {

	os.RemoveAll(root)
	os.MkdirAll(root, 0777)

	for i := 0; i < count; i++ {
		f, err := os.Create(root + "/" + strconv.Itoa(i))

		if err != nil {
			panic(err)
		}

		rs := randStringRunes(1000000)
		f.WriteString(rs)
	}
}
