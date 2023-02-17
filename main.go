package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	. "github.com/jorgebay/polar-benchmark-tool/internal"
)

var lastError atomic.Value

// Based on h2load parameter names
var requestsLength = flag.Int("n", 100, "Number of  requests across all  clients")
var clientsLength = flag.Int("c", 1, "Number of clients")
var maxConcurrentStreams = flag.Int("m", 32, "Max concurrent requests to issue per client.")
var maxConnectionsPerHost = flag.Int("ch", 16, "For HTTP/1.1, determines the max connections per host.")
var hosts = flag.String("hosts", "", "The host addresses of the endpoint(s)")
var workloadName = flag.String("w", "default", "The name of the workload (binary, get, default)")
var messagesPerRequest = flag.Int("mr", 16, "Number of messages per request in the workload (when supported)")
var useH2 = flag.Bool("h2", false, "For workloads that allow it, use HTTP/2")

var histogram = hdrhistogram.New(1, 4_000_000, 4)

func main() {
	flag.Parse()

	if *hosts == "" {
		panic("Host addresses are required, e.g. 10.0.0.100")
	}

	fmt.Printf("Starting benchmark. %d total client(s). %d total requests\n", *clientsLength, *requestsLength)

	workload := BuildWorkload(*workloadName, *hosts, *messagesPerRequest)
	fmt.Println("Initializing")
	workload.Init()
	protocolInfo := ""
	if *useH2 {
		protocolInfo = HTTP2
	}

	fmt.Println("Warming up")
	warmup(workload, protocolInfo)

	fmt.Println("Starting workload", *workloadName)
	totalResponses := int64(0)
	var wg sync.WaitGroup
	requestsPerClient := *requestsLength / *clientsLength
	start := time.Now()
	for i := 0; i < *clientsLength; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			okResponses := runClient(requestsPerClient, *maxConcurrentStreams, workload, protocolInfo, false)
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
		float64(histogram.ValueAtQuantile(50))/1_000,
		float64(histogram.ValueAtQuantile(99.9))/1_000,
		float64(histogram.ValueAtQuantile(100))/1_000)
	reqThroughput := (totalResponses * 1_000_000) / timeSpent.Microseconds()
	fmt.Printf(
		"Throughput %d messages/s (%d req/s)\n",
		reqThroughput*int64(workload.MessagesPerPayload()),
		reqThroughput)
}

func warmup(workload Workload, protocolInfo string) {
	runClient(10000, 16, workload, protocolInfo, true)
}

func runClient(requestsLength int, maxConcurrentStreams int, workload Workload, protocolInfo string, isWarmup bool) int64 {
	client := workload.NewClient(*maxConnectionsPerHost, protocolInfo)
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

func doRequest(client WorkloadClient, w Workload, v int, isWarmup bool) bool {
	start := time.Now()
	err := client.DoRequest(v)
	if err != nil {
		lastError.Store(err.Error())
		return false
	}
	if !isWarmup {
		histogram.RecordValue(time.Since(start).Microseconds())
	}
	return true
}
