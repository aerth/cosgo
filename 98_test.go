package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var times []time.Duration

func init() {
	rand.NewSource(time.Now().UnixNano())
	times = randomDuration()
}
func randomDuration() (times []time.Duration) {
	t := uint64(rand.Uint32()*4) * uint64(rand.Uint32()) * uint64(rand.Uint32()*rand.Uint32()) / uint64(rand.Intn(10000)+1)
	times = append(times, time.Duration(t))
	return times
}

func BenchmarkHumanize(b *testing.B) {
	for i := 0; i < 1000; i++ {
		rands := randomDuration()
		for _, t := range rands {
			b.StartTimer()
			humanize(t)
			b.StopTimer()
		}
	}
}

func TestHumanize(t *testing.T) {
	s := humanize(time.Hour * 24 * 7 * 30)
	fmt.Println("Got:", s)
	fmt.Println("Wanted:", "8 months ago")
	if s != "8 months ago" {
		t.Fail()
	}
	// Output: 8 months ago
}
