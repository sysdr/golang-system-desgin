package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"
)

const (
	numRequests      = 100000 // Number of requests for local benchmarking
	targetRPS        = 100000000 // 100 million RPS
	defaultDelayUs   = 0     // Default artificial delay in microseconds
)

// simpleHandler is a very basic HTTP handler that just writes "OK"
// It can optionally include an artificial delay.
func simpleHandler(w http.ResponseWriter, r *http.Request) {
	delayStr := r.URL.Query().Get("delay_us")
	delayUs, err := strconv.Atoi(delayStr)
	if err != nil || delayUs < 0 {
		delayUs = defaultDelayUs
	}

	if delayUs > 0 {
		time.Sleep(time.Duration(delayUs) * time.Microsecond)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func main() {
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("🚀 Quantitative Reality of 100M RPS Calculator 🚀")
	fmt.Println("--------------------------------------------------------------------------------")

	// Create an in-memory HTTP test server
	testServer := httptest.NewServer(http.HandlerFunc(simpleHandler))
	defer testServer.Close()

	fmt.Printf("Benchmarking a single, minimal HTTP handler (%d requests)...\n", numRequests)
	fmt.Printf("Handler URL: %s\n", testServer.URL)
	fmt.Printf("Default artificial delay per request: %d microseconds\n", defaultDelayUs)

	// --- Benchmark with default delay ---
	fmt.Println("\n--- Baseline Measurement (no extra delay) ---")
	measureAndReport(testServer.URL, defaultDelayUs)

	// --- Assignment: Benchmark with 10us delay ---
	fmt.Println("\n--- Assignment Measurement (10us artificial delay) ---")
	measureAndReport(testServer.URL, 10)

	// --- Assignment: Benchmark with 50us delay ---
	fmt.Println("\n--- Assignment Measurement (50us artificial delay) ---")
	measureAndReport(testServer.URL, 50)

	// --- Assignment: Benchmark with 100us delay ---
	fmt.Println("\n--- Assignment Measurement (100us artificial delay) ---")
	measureAndReport(testServer.URL, 100)


	fmt.Println("\n--------------------------------------------------------------------------------")
	fmt.Println("💡 Insights for 100M RPS: Small latencies have huge consequences!")
	fmt.Println("--------------------------------------------------------------------------------")
}

func measureAndReport(url string, delayUs int) {
	client := &http.Client{}

	start := time.Now()
	var wg sync.WaitGroup
	// Simulate concurrent requests to better reflect real-world load
	// Note: httptest.Server is single-threaded, but this simulates client-side concurrency
	// For actual server-side concurrency testing, use a real server and a load testing tool.
	concurrency := 100 // A reasonable number of concurrent clients for httptest
	requestsPerGoroutine := numRequests / concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				reqURL := fmt.Sprintf("%s?delay_us=%d", url, delayUs)
				resp, err := client.Get(reqURL)
				if err != nil {
					// In a real system, you'd handle this more robustly
					// For this benchmark, we'll just print and continue
					// fmt.Printf("Error making request: %vn", err)
					continue
				}
				resp.Body.Close() // Important to close the body to prevent resource leaks
			}
		}()
	}
	wg.Wait()

	elapsed := time.Since(start)
	observedRPS := float64(numRequests) / elapsed.Seconds()
	instancesNeeded := targetRPS / observedRPS

	fmt.Printf("  - Artificial Delay: %d microseconds\n", delayUs)
	fmt.Printf("  - Total requests processed: %d\n", numRequests)
	fmt.Printf("  - Total time taken: %s\n", elapsed)
	fmt.Printf("  - Observed RPS (single instance): %.2f req/s\n", observedRPS)
	fmt.Printf("  - Instances needed for 100M RPS: %.2f instances\n", instancesNeeded)
	fmt.Printf("  - Cost per request (avg): %.2f ns\n", float64(elapsed.Nanoseconds()) / float64(numRequests))
}

