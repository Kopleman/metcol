package utils

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFanIn(t *testing.T) {
	t.Run("multiple channels with data", func(t *testing.T) {
		ch1 := make(chan int)
		ch2 := make(chan int)
		ch3 := make(chan int)

		go func() {
			defer close(ch1)
			ch1 <- 1
			ch1 <- 2
		}()

		go func() {
			defer close(ch2)
			ch2 <- 3
			ch2 <- 4
		}()

		go func() {
			defer close(ch3)
			ch3 <- 5
			ch3 <- 6
		}()

		out := FanIn(ch1, ch2, ch3)

		var results []int
		for v := range out {
			results = append(results, v)
		}

		assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6}, results)
	})

	t.Run("single channel", func(t *testing.T) {
		ch := make(chan string)
		go func() {
			defer close(ch)
			ch <- "foo"
			ch <- "bar"
		}()

		out := FanIn(ch)

		var results []string
		for v := range out {
			results = append(results, v)
		}

		assert.ElementsMatch(t, []string{"foo", "bar"}, results)
	})

	t.Run("no input channels", func(t *testing.T) {

		out := FanIn[int]()

		select {
		case _, ok := <-out:
			assert.False(t, ok, "channel should be closed")
		default:
			t.Error("channel should be closed immediately")
		}
	})

	t.Run("closed channels", func(t *testing.T) {
		ch1 := make(chan int)
		close(ch1)

		ch2 := make(chan int)
		close(ch2)

		out := FanIn(ch1, ch2)

		select {
		case _, ok := <-out:
			assert.False(t, ok, "channel should be closed")
		case <-time.After(100 * time.Millisecond):
			t.Error("channel should be closed immediately")
		}
	})

	t.Run("concurrent safety", func(t *testing.T) {
		// Arrange
		const numChannels = 10
		const itemsPerChannel = 100

		var chs []chan int
		var wg sync.WaitGroup

		// Create channels and start writers
		for i := 0; i < numChannels; i++ {
			ch := make(chan int)
			chs = append(chs, ch)

			wg.Add(1)
			go func(ch chan int) {
				defer wg.Done()
				defer close(ch)
				for j := 0; j < itemsPerChannel; j++ {
					ch <- j
				}
			}(ch)
		}

		// Act
		out := FanIn(chs...)

		// Collect results
		var results []int
		var outWg sync.WaitGroup
		outWg.Add(1)
		go func() {
			defer outWg.Done()
			for v := range out {
				results = append(results, v)
			}
		}()

		// Wait for all writers and close out channel
		wg.Wait()
		outWg.Wait()

		// Assert
		assert.Len(t, results, numChannels*itemsPerChannel)
	})
}
