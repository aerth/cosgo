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
	t := uint64(rand.Uint32()*4) * uint64(rand.Uint32()) * uint64(rand.Uint32()*rand.Uint32()) / 100000
	times = append(times, time.Duration(t))
	return times
}
func TestOne(t *testing.T) {
	fmt.Println("hello")
	// Output: ohno
}

func BenchmarkHumanize(b *testing.B) {
	for i := 0; i < 10; i++ {
		rands := randomDuration()
		for _, t := range rands {
			fmt.Println(humanize(t))
			// Output: unknown
		}
	}
}

func TestHumanize(t *testing.T) {
	fmt.Println(humanize(time.Hour * 24 * 7 * 30))
	// Output: 8 months ag
}
