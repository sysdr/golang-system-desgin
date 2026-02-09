package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"heap_cost_analyzer/internal/analyzer"
)

var (
	heapRequestCount  uint64
	stackRequestCount uint64
	lastHeapStats    *analyzer.RequestStats
	lastStackStats   analyzer.RequestStats
	metricsMu        sync.RWMutex
)

func main() {
	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/stats-heap", func(w http.ResponseWriter, r *http.Request) {
		reqIDStr := r.URL.Query().Get("id")
		reqID, err := strconv.ParseUint(reqIDStr, 10, 64)
		if err != nil {
			reqID = 0
		}
		stats := analyzer.ProcessRequestPointer(reqID)
		metricsMu.Lock()
		heapRequestCount++
		lastHeapStats = stats
		metricsMu.Unlock()
		fmt.Fprintf(w, "Heap Allocated Stats: %+v\n", stats)
		log.Printf("Heap handler processed request ID %d", reqID)
		runtime.GC()
	})

	http.HandleFunc("/stats-stack", func(w http.ResponseWriter, r *http.Request) {
		reqIDStr := r.URL.Query().Get("id")
		reqID, err := strconv.ParseUint(reqIDStr, 10, 64)
		if err != nil {
			reqID = 0
		}
		stats := analyzer.ProcessRequestValue(reqID)
		metricsMu.Lock()
		stackRequestCount++
		lastStackStats = stats
		metricsMu.Unlock()
		fmt.Fprintf(w, "Stack Allocated Stats: %+v\n", stats)
		log.Printf("Stack handler processed request ID %d", reqID)
		runtime.GC()
	})

	http.HandleFunc("/metrics", metricsJSONHandler)

	log.Printf("Server listening on :%s", "8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	metricsMu.RLock()
	heapCnt := heapRequestCount
	stackCnt := stackRequestCount
	lastHeap := lastHeapStats
	lastStack := lastStackStats
	metricsMu.RUnlock()

	heapID, heapTS, heapDur, heapStatus := uint64(0), int64(0), time.Duration(0), 0
	if lastHeap != nil {
		heapID, heapTS, heapDur, heapStatus = lastHeap.ID, lastHeap.Timestamp, lastHeap.Duration, lastHeap.Status
	}
	stackID, stackTS, stackDur, stackStatus := lastStack.ID, lastStack.Timestamp, lastStack.Duration, lastStack.Status

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Heap Cost Analyzer Dashboard</title>
<meta http-equiv="refresh" content="2">
<style>
body{font-family:sans-serif;margin:20px;background:#1a1a2e;color:#eee;}
h1{color:#0f4;}
.metric{background:#16213e;padding:12px;margin:8px 0;border-radius:8px;}
.metric strong{color:#0f4;}
.nav a{color:#0f4;margin-right:12px;}
</style>
</head>
<body>
<h1>Heap Cost Analyzer Dashboard</h1>
<p class="nav"><a href="/">Dashboard</a> | <a href="/stats-heap?id=1">Demo Heap</a> | <a href="/stats-stack?id=2">Demo Stack</a> | <a href="/metrics">JSON Metrics</a></p>
<div class="metric"><strong>Heap requests:</strong> %d</div>
<div class="metric"><strong>Stack requests:</strong> %d</div>
<div class="metric"><strong>Last heap stats:</strong> ID=%d Timestamp=%d Duration=%s Status=%d</div>
<div class="metric"><strong>Last stack stats:</strong> ID=%d Timestamp=%d Duration=%s Status=%d</div>
<p><em>Auto-refresh 2s. Hit Demo links to update metrics.</em></p>
</body>
</html>`,
		heapCnt, stackCnt,
		heapID, heapTS, heapDur, heapStatus,
		stackID, stackTS, stackDur, stackStatus)
	fmt.Fprint(w, html)
}

func metricsJSONHandler(w http.ResponseWriter, r *http.Request) {
	metricsMu.RLock()
	defer metricsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if lastHeapStats != nil {
		fmt.Fprintf(w, `{"heap_requests":%d,"stack_requests":%d,"last_heap":%+v,"last_stack":%+v}`,
			heapRequestCount, stackRequestCount, lastHeapStats, lastStackStats)
	} else {
		fmt.Fprintf(w, `{"heap_requests":%d,"stack_requests":%d,"last_stack":%+v}`,
			heapRequestCount, stackRequestCount, lastStackStats)
	}
}
