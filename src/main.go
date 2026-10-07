package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	client     *http.Client
	timeoutDur time.Duration
	suc, fail  atomic.Int64
)

func randomInt(min, max int) int {
	if max < min {
		return min
	}
	return rand.IntN(max-min+1) + min
}

func generateIP() string {
	class := randomInt(1, 3)

	var firstOctet uint8
	switch class {
	case 1:
		firstOctet = uint8(randomInt(1, 126))
	case 2:
		firstOctet = uint8(randomInt(128, 191))
	case 3:
		firstOctet = uint8(randomInt(192, 223))
	}

	secondOctet := uint8(randomInt(0, 255))
	thirdOctet := uint8(randomInt(1, 255))
	fourthOctet := uint8(randomInt(1, 254))

	return fmt.Sprintf("%d.%d.%d.%d", firstOctet, secondOctet, thirdOctet, fourthOctet)
}

//go:embed user_agents.txt
var userAgentsRaw string

var userAgents []string

func init() {
	lines := strings.Split(strings.TrimSpace(userAgentsRaw), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			userAgents = append(userAgents, line)
		}
	}
	if len(userAgents) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: embedded user_agents.txt is empty")
	}
}

func fetchUrl(url string, mask bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDur)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request for URL %s: %v\n", url, err)
		updateStats(0)
		return
	}
	req.Close = true

	if mask {
		if len(userAgents) == 0 {
			fmt.Println("No user agents available, skipping mask")
			updateStats(0)
			return
		}
		randomIndex := randomInt(0, len(userAgents)-1)

		fIP := generateIP()
		req.Header.Set("Forwarded", fmt.Sprintf("for=%s; proto=https", fIP))
		req.Header.Set("X-Forwarded-For", fIP)
		req.Header.Set("User-Agent", userAgents[randomIndex])
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:156.0) Gecko/20100101 Firefox/156.0")
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("Request to %s timed out after %d seconds\n", url, int(timeoutDur.Seconds()))
			updateStats(503)
			return
		}
		fmt.Printf("Error fetching URL %s: %v\n", url, err)
		updateStats(404)
		return
	}
	defer resp.Body.Close()

	updateStats(resp.StatusCode)
	return
}

func randomizedRamp(numRequests int, sem chan struct{}, wg *sync.WaitGroup, url string, mask bool) {
	remReqs := numRequests
	divisor := float64(randomInt(1100, 1600)) / 1000.0
	minBatchBase := int(float64(numRequests) / divisor)

	for remReqs > 0 {
		randMin := randomInt(6400, 10000)
		minBatch := randMin
		if minBatchBase > randMin {
			minBatch = minBatchBase
		}
		if minBatch > remReqs {
			minBatch = remReqs
		}

		batchSize := randomInt(minBatch, remReqs)
		for i := 0; i < batchSize; i++ {
			sem <- struct{}{}
			wg.Go(func() {
				defer func() { <-sem }()
				fetchUrl(url, mask)
			})
		}
		remReqs -= batchSize

		time.Sleep(time.Duration(randomInt(128, 640)) * time.Millisecond)
	}
}

func updateStats(statCode int) {
	if statCode >= 200 && statCode < 300 {
		suc.Add(1)
	} else {
		fail.Add(1)
	}
}

func printStat() {
	s := suc.Load()
	f := fail.Load()
	fmt.Printf("\nSuccessful requests: %d\nUnsuccessful requests: %d\n", s, f)
	fmt.Printf("Total requests: %d\n", s+f)
}

func main() {
	urlPtr := flag.String("url", "", "Target URL (required)")
	flag.StringVar(urlPtr, "u", "", "shorthand for -url")

	requestsPtr := flag.Int("requests", 1, "Number of requests (default: random 20-32)")
	flag.IntVar(requestsPtr, "r", 1, "shorthand for -requests")

	concrtPtr := flag.Int("concurrency", 512, "Maximum number of concurrent requests (default: 512)")
	flag.IntVar(concrtPtr, "c", 512, "shorthand for -concurrency")

	rampPtr := flag.Bool("ranramp", false, "Enable randomized request ramp-up (default: false)")
	maskPtr := flag.Bool("mask", true, "Use IP masking (default: true)")

	runsPtr := flag.Int("runs", 1, "Number of runs to execute (default: 1)")
	runSleepPtr := flag.Duration("run-sleep", 10*time.Second, "Sleep duration between runs (default: 10s)")

	printUsage := func(printReceivedArgs bool) {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -url https://example.com -ranramp\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -url https://example.com -requests 50 -mask false\n", os.Args[0])

		if printReceivedArgs {
			fmt.Fprintf(os.Stderr, "\nReceived arguments:\n")
			for i, arg := range os.Args {
				fmt.Fprintf(os.Stderr, "  [%d] %s\n", i, arg)
			}
		}
	}

	flag.Usage = func() { printUsage(false) }

	args := os.Args[:1]
	for _, a := range os.Args[1:] {
		if !strings.HasPrefix(a, "/proc/self/fd/") {
			args = append(args, a)
		}
	}
	os.Args = args

	flag.Parse()

	if *urlPtr == "" {
		fmt.Fprintln(os.Stderr, "Error: Target URL is required.")

		printUsage(true)
		os.Exit(1)
	}
	url := *urlPtr
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	numRequests := *requestsPtr
	if numRequests <= 0 {
		numRequests = randomInt(20, 32)
		fmt.Printf("Using random number of requests: %d\n", numRequests)
	}
	if numRequests > 1000000 {
		fmt.Printf("Cannot create more than 1000000 requests, given %d requests\nCreating 1000000 requests\n", numRequests)
		numRequests = 1000000
	}

	maxConcurrent := *concrtPtr
	if maxConcurrent > 10000 {
		fmt.Fprintf(os.Stderr, "Warning: Capping concurrency at 10000 (requested %d)\n", maxConcurrent)
		maxConcurrent = 10000
	}

	timeoutDur = time.Duration(randomInt(32, 64)) * time.Second

	client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: 0,
			IdleConnTimeout:     1 * time.Millisecond,
			TLSHandshakeTimeout: 10 * time.Second,
			ForceAttemptHTTP2:   false,
		},
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		printStat()
		fmt.Println("Requests interrupted")
		os.Exit(0)
	}()

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for runNum := 1; runNum <= *runsPtr; runNum++ {
		if *runsPtr > 1 {
			fmt.Printf("\n--- Run %d/%d ---\n", runNum, *runsPtr)
		}

		suc.Store(0)
		fail.Store(0)

		fmt.Printf("Starting %d requests to %s (Max concurrent: %d), use ctrl+c to stop...\n", numRequests, url, maxConcurrent)

		if *rampPtr {
			randomizedRamp(numRequests, sem, &wg, url, *maskPtr)
		} else {
			for i := 0; i < numRequests; i++ {
				sem <- struct{}{}
				wg.Go(func() {
					defer func() { <-sem }()
					fetchUrl(url, *maskPtr)
				})
			}
		}

		wg.Wait()
		fmt.Print("\n")
		printStat()

		if runNum < *runsPtr {
			fmt.Printf("Sleeping for %v before next run...\n", *runSleepPtr)
			time.Sleep(*runSleepPtr)
		}
	}
	fmt.Println("\nAll runs completed")
}
