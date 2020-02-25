package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"golang.org/x/sync/semaphore"
)

var words = []string{"one", "two", "three", "four", "five", "six", "seven", "eight"}

func main() {
	err := logStringsInList(words)
	if err != nil {
		log.Fatal(err)
	}
}

// logStringsInList takes a slice of strings and logs them to the console
func logStringsInList(list []string) error {
	errChan := make(chan error, 1) // create new error channel and wait group
	var wg sync.WaitGroup
	wg.Add(len(list))
	for _, word := range list {
		go printer(os.Stdout, word, &wg, errChan) // print each 'word' using its own concurrent go routine
	}
	wg.Wait()
	close(errChan)
	return <-errChan
}

// printer takes a writer, string, wait group and an error channel. It writes the string using the provided writer.
// if an error occurs, it is sent back over the error channel. If another printer (that is running concurrently to this one)
// has already errored, this printers error will be ignored. The number of concurrent goroutines is limited by the semaphore
func printer(out io.Writer, w string, wg *sync.WaitGroup, errChan chan error) {
	if wg != nil {
		defer wg.Done() // make sure wait group closes
	}
	sem := semaphore.NewWeighted(int64(4)) // limit the number of goroutines running at any one time with the use of semaphores
	ctx := context.Background()
	if err := sem.Acquire(ctx, 1); err != nil {
		select {
		case errChan <- err: // channel not blocked
		default: // channel is blocked
		}
	}
	if _, err := out.Write([]byte(fmt.Sprint(w, "\n"))); err != nil {
		select {
		case errChan <- err: // not blocked
		default: // blocked
		}
	}
	sem.Release(1)
}
