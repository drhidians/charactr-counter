package main

import "testing"

func TestCharH(t *testing.T) {
	_, err := CharacterHistogram("files")
	if err != nil {
		t.Fatalf("%s", err)
	}
}
func BenchmarkCharH(b *testing.B) {
	for n := 0; n < b.N; n++ {
		CharacterHistogram("files")
	}
}

// go test -bench=. -run=XXX
