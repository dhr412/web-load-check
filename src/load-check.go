package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var client *http.Client
var timeoutDur time.Duration = time.Second * 64

func fetchpag(urlStr string, wg *sync.WaitGroup, mask bool) int {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDur)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		fmt.Printf("Error creating request for URL %s: %v\n", urlStr, err)
		return 0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0")
	if mask {
		fIP := generateIP()
		req.Header.Set("Forwarded", fmt.Sprintf("for=%s; proto=https", fIP))
		req.Header.Set("X-Forwarded-For", fIP)
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("Request to %s timed out after %d seconds\n", urlStr, int(timeoutDur))
			return 503
		}
		fmt.Printf("Error fetching URL %s: %v\n", urlStr, err)
		return 404
	}
	defer resp.Body.Close()
	return resp.StatusCode
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
	thirdOctet := randomInt(0, 255)
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

func main() {
	urlPtr := flag.String("url", "", "Target URL (required)")
	requestsPtr := flag.Int("requests", 1, "Number of requests (default: random 20-32)")
	maskPtr := flag.Bool("mask", true, "Use IP masking (default: true)")
	helpPtr := flag.Bool("help", false, "Show help message")
	flag.Parse()
	if *helpPtr || *urlPtr == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n  %s -url https://example.com\n  %s -url https://example.com -requests 50 -mask=false\n", os.Args[0], os.Args[0])
		os.Exit(0)
	}
	client = &http.Client{
		Timeout: timeoutDur,
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
	mask := *maskPtr
	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		res []int
	)
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			code := fetchpag(url, &wg, mask)
			mu.Lock()
			res = append(res, code)
			mu.Unlock()
		}()
	}
	wg.Wait()
	suc, fail := 0, 0
	for _, code := range res {
		if code >= 200 && code < 300 {
			suc++
		} else {
			fail++
		}
	}
	fmt.Printf("Successful requests: %d\nUnsuccessful requests: %d\n", suc, fail)
	fmt.Println("Requests completed")
}
