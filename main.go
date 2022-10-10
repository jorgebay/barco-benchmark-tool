package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/jorgebay/barco-benchmark-tool/internal"
	"golang.org/x/net/http2"
)

var lastError atomic.Value

func main() {
	// Based on h2load parameter names
	requestsLength := flag.Int("n", 100, "Number of  requests across all  clients")
	clientsLength := flag.Int("c", 1, "Number of clients")
	maxConcurrentStreams := flag.Int("m", 1024, "Max  concurrent  requests  to issue  per  session")
	url := flag.String("u", "", "The uri of the endpoint")

    flag.Parse()

	if *url == "" {
		panic("Uri is required")
	}

	fmt.Printf("Starting benchmark. %d total client(s). %d total requests", *clientsLength, *requestsLength)

	workload := NewRandomWorkload(*url)
	workload.Init()

	fmt.Println("Warming up")
	warmup(workload)

	fmt.Println("Starting workload")
	totalResponses := int64(0)
	var wg sync.WaitGroup
	requestsPerClient := *requestsLength / *clientsLength
	start := time.Now()
	for i := 0; i < *clientsLength; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			okResponses := runClient(requestsPerClient, *maxConcurrentStreams, workload)
			atomic.AddInt64(&totalResponses, int64(okResponses))
		}()
	}
	wg.Wait()

	fmt.Printf("Finished. Total responses %d in %dms\n", atomic.LoadInt64(&totalResponses), time.Since(start).Milliseconds())

	totalErrors := int64(requestsPerClient * *clientsLength) - atomic.LoadInt64(&totalResponses)
	if totalErrors > 0 {
		fmt.Printf("Encountered %d errors", totalErrors)

		errMessage := lastError.Load()
		if errMessage != nil {
			fmt.Printf(". Last error: %s", errMessage)
		}
		fmt.Println()
	}
}

func warmup(workload Workload) {
	runClient(1000, 16, workload)
}

func runClient(requestsLength int, maxConcurrentStreams int, workload Workload) int64 {
	client := http.Client{

		Transport: &http2.Transport{
			StrictMaxConcurrentStreams: true,
			AllowHTTP:                  true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				// Pretend we are dialing a TLS endpoint.
				return net.Dial(network, addr)
			},
			ReadIdleTimeout: 1 * time.Second,
		},
	}

	c := make(chan bool, maxConcurrentStreams)
	for i := 0; i < maxConcurrentStreams; i++ {
		c <- true
	}

	counter := int64(0)
	startIndex := rand.Intn(1<<31)

	for i := 0; i < requestsLength; i++ {
		<- c
		go func(v int) {
			success := doRequest(client, workload, startIndex+v)
			c <- true
			if success {
				atomic.AddInt64(&counter, 1)
			}
		}(i)
	}

	// Receive the last ones
	for i := 0; i < maxConcurrentStreams; i++ {
		<- c
	}

	return atomic.LoadInt64(&counter)
}

func doRequest(client http.Client, w Workload, v int) bool {
	req, err := http.NewRequest(w.Method(), w.Url(), w.Body(v))
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		lastError.Store(err.Error())
		return false
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		lastError.Store(string(body))
		return false
	}
	return true
}
