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

	"github.com/HdrHistogram/hdrhistogram-go"
	. "github.com/jorgebay/barco-benchmark-tool/internal"
	"golang.org/x/net/http2"
)

type httpVersion int

const (
	h1 httpVersion = iota
	h2
)

var lastError atomic.Value

// Based on h2load parameter names
var requestsLength = flag.Int("n", 100, "Number of  requests across all  clients")
var clientsLength = flag.Int("c", 1, "Number of clients")
var maxConcurrentStreams = flag.Int("m", 32, "Max concurrent requests to issue per client.")
var url = flag.String("u", "", "The uri(s) of the endpoint(s)")
var workloadName = flag.String("w", "default", "The name of the workload")
var messagesPerRequest = flag.Int("mr", 16, "Number of messages per request in the workload (when supported)")
var useH1 = flag.Bool("h1", false, "Force http/1.1")

var histogram = hdrhistogram.New(1, 4_000_000, 4)

func main() {
	flag.Parse()

	if *url == "" {
		panic("Uri is required")
	}

	fmt.Printf("Starting benchmark. %d total client(s). %d total requests\n", *clientsLength, *requestsLength)

	workload := BuildWorkload(*workloadName, *url, *messagesPerRequest)
	fmt.Println("Initializing")
	workload.Init()
	version := h2
	if *useH1 {
		version = h1
	}

	fmt.Println("Warming up")
	warmup(workload, version)

	fmt.Println("Starting workload", *workloadName)
	totalResponses := int64(0)
	var wg sync.WaitGroup
	requestsPerClient := *requestsLength / *clientsLength
	start := time.Now()
	for i := 0; i < *clientsLength; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			okResponses := runClient(requestsPerClient, *maxConcurrentStreams, workload, version, false)
			atomic.AddInt64(&totalResponses, int64(okResponses))
		}()
	}
	wg.Wait()

	printResult(start, atomic.LoadInt64(&totalResponses), workload)
}

func printResult(start time.Time, totalResponses int64, workload Workload) {
	timeSpent := time.Since(start)
	fmt.Printf("Finished. Total responses %d in %dms.\n", totalResponses, timeSpent.Milliseconds())

	requestsPerClient := *requestsLength / *clientsLength
	totalErrors := int64(requestsPerClient**clientsLength) - totalResponses

	if totalErrors > 0 {
		fmt.Printf("Encountered %d errors", totalErrors)

		errMessage := lastError.Load()
		if errMessage != nil {
			fmt.Printf(". Last error: %s", errMessage)
		}
		fmt.Println()
		return
	}

	fmt.Printf(
		"Latency in ms p50: %.1f; p999: %.1f max: %.1f.\n",
		float64(histogram.ValueAtQuantile(50)) / 1_000,
		float64(histogram.ValueAtQuantile(99.9)) / 1_000,
		float64(histogram.ValueAtQuantile(100)) / 1_000)
	reqThroughput := (totalResponses * 1_000_000) / timeSpent.Microseconds()
	fmt.Printf(
		"Throughput %d messages/s (%d req/s)\n",
		reqThroughput*int64(workload.MessagesPerPayload()),
		reqThroughput)
}

func warmup(workload Workload, v httpVersion) {
	runClient(10000, 16, workload, v, true)
}

func runClient(requestsLength int, maxConcurrentStreams int, workload Workload, v httpVersion, isWarmup bool) int64 {
	var transport http.RoundTripper

	if v == h1 {
		transport = &http.Transport{
			ForceAttemptHTTP2:     false,
			MaxIdleConns:          100,
			IdleConnTimeout:       10 * time.Second,
		}
	} else {
		transport = &http2.Transport{
			StrictMaxConcurrentStreams: true,
			AllowHTTP:                  true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				// Pretend we are dialing a TLS endpoint.
				return net.Dial(network, addr)
			},
			ReadIdleTimeout: 1 * time.Second,
		}
	}

	client := http.Client{
		Transport: transport,
	}

	c := make(chan bool, maxConcurrentStreams)
	for i := 0; i < maxConcurrentStreams; i++ {
		c <- true
	}

	counter := int64(0)
	startIndex := rand.Intn(1 << 31)

	for i := 0; i < requestsLength; i++ {
		<-c
		go func(v int) {
			success := doRequest(client, workload, startIndex+v, isWarmup)
			c <- true
			if success {
				atomic.AddInt64(&counter, 1)
			}
		}(i)
	}

	// Receive the last ones
	for i := 0; i < maxConcurrentStreams; i++ {
		<-c
	}

	return atomic.LoadInt64(&counter)
}

func doRequest(client http.Client, w Workload, v int, isWarmup bool) bool {
	req, err := http.NewRequest(w.Method(), w.Url(), w.Body(v))
	if w.ContentType() != "" {
		req.Header.Add("Content-Type", w.ContentType())
	}
	if err != nil {
		panic(err)
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		lastError.Store(err.Error())
		return false
	}

	defer resp.Body.Close()
	if !isWarmup {
		histogram.RecordValue(time.Since(start).Microseconds())
	}

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
