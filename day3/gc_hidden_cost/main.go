package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"gc_hidden_cost/processor"
)

// Global processors for demonstration
var naiveProcessor *processor.NaiveProcessor
var pooledProcessor *processor.PooledProcessor

// Live stats for dashboard (request counts and last duration per processor type)
var (
	statsMu                    sync.RWMutex
	naiveRequestCount           uint64
	pooledRequestCount          uint64
	lastNaiveDurationNs         int64
	lastPooledDurationNs        int64
)

func init() {
	naiveProcessor = processor.NewNaiveProcessor()
	pooledProcessor = processor.NewPooledProcessor()
}

// NewMux returns the application's HTTP mux for testing and reuse.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboardHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/naive", handleProcessWithStats(naiveProcessor, recordNaiveStats))
	mux.HandleFunc("/pooled", handleProcessWithStats(pooledProcessor, recordPooledStats))
	mux.HandleFunc("/debug/mem", handleMemStats)
	mux.HandleFunc("/api/stats", handleStatsJSON)
	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s...", addr)
	log.Fatal(http.ListenAndServe(addr, NewMux()))
}

func handleProcess(p processor.Processor) http.HandlerFunc {
	return handleProcessWithStats(p, nil)
}

func handleProcessWithStats(p processor.Processor, recordStats func(time.Duration)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sizeStr := r.URL.Query().Get("size")
		if sizeStr == "" {
			http.Error(w, "missing 'size' parameter", http.StatusBadRequest)
			return
		}

		size, err := strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			http.Error(w, "invalid 'size' parameter", http.StatusBadRequest)
			return
		}

		start := time.Now()
		result, err := p.Process(size)
		duration := time.Since(start)
		if err != nil {
			http.Error(w, fmt.Sprintf("processing error: %v", err), http.StatusInternalServerError)
			return
		}
		if recordStats != nil {
			recordStats(duration)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "success",
			"message":  result,
			"duration": duration.String(),
			"processor": fmt.Sprintf("%T", p),
		})
	}
}

func recordNaiveStats(d time.Duration) {
	statsMu.Lock()
	naiveRequestCount++
	lastNaiveDurationNs = d.Nanoseconds()
	statsMu.Unlock()
}

func recordPooledStats(d time.Duration) {
	statsMu.Lock()
	pooledRequestCount++
	lastPooledDurationNs = d.Nanoseconds()
	statsMu.Unlock()
}

func handleStatsJSON(w http.ResponseWriter, r *http.Request) {
	statsMu.RLock()
	defer statsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"naive_request_count":    naiveRequestCount,
		"pooled_request_count":   pooledRequestCount,
		"last_naive_duration_ns": lastNaiveDurationNs,
		"last_pooled_duration_ns": lastPooledDurationNs,
	})
}

func handleMemStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Alloc":        m.Alloc,        // bytes allocated and still in use
		"TotalAlloc":   m.TotalAlloc,   // bytes allocated (even if freed)
		"Sys":          m.Sys,          // total bytes obtained from OS
		"NumGC":        m.NumGC,        // number of completed GC cycles
		"GCCPUFraction":m.GCCPUFraction,// fraction of CPU time spent in GC
		"PauseTotalNs": m.PauseTotalNs, // total GC pause time in nanoseconds
		"HeapObjects":  m.HeapObjects,  // # objects in the heap
		"LastGC":       time.Unix(0, int64(m.LastGC)).Format(time.RFC3339Nano),
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	lastGCStr := time.Unix(0, int64(m.LastGC)).Format(time.RFC3339Nano)

	statsMu.RLock()
	naiveCnt := naiveRequestCount
	pooledCnt := pooledRequestCount
	lastNaiveNs := lastNaiveDurationNs
	lastPooledNs := lastPooledDurationNs
	statsMu.RUnlock()

	lastNaiveStr := "—"
	if lastNaiveNs > 0 {
		lastNaiveStr = time.Duration(lastNaiveNs).String()
	}
	lastPooledStr := "—"
	if lastPooledNs > 0 {
		lastPooledStr = time.Duration(lastPooledNs).String()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>GC Hidden Cost Dashboard</title>
<meta http-equiv="refresh" content="2">
<style>
body{font-family:system-ui,sans-serif;margin:24px;background:#f8f9fa;color:#212529;}
h1{color:#212529;font-weight:600;}
h2{color:#495057;font-size:1.1rem;margin-top:24px;}
.info{background:#fff;border:1px solid #dee2e6;border-radius:8px;padding:16px;margin:12px 0;color:#495057;line-height:1.5;}
.metric{background:#fff;border:1px solid #dee2e6;padding:12px 16px;margin:8px 0;border-radius:8px;}
.metric strong{color:#212529;}
.nav{margin:16px 0;}
.nav a{color:#495057;margin-right:16px;text-decoration:none;}
.nav a:hover{text-decoration:underline;}
.live{background:#fff3cd;border:1px solid #ffc107;color:#856404;}
.actions{margin:16px 0;}
.actions a{display:inline-block;margin-right:8px;padding:8px 16px;background:#212529;color:#fff;border-radius:6px;text-decoration:none;}
.actions a:hover{background:#495057;}
</style>
</head>
<body>
<h1>GC Hidden Cost Dashboard</h1>
<div class="info">
<strong>About this project:</strong> This app compares two ways of handling per-request buffers in Go: <b>Naive</b> (new allocation every time) vs <b>Pooled</b> (sync.Pool reuse). Use the actions below to generate load; watch GC metrics and live request stats to see how pooling reduces allocations and GC pressure.
</div>
<p class="nav"><a href="/">Dashboard</a> | <a href="/debug/mem">JSON MemStats</a> | <a href="/api/stats">JSON Live Stats</a></p>

<h2>Live request metrics</h2>
<p>Click the buttons to run requests; counts and last duration update in real time (page auto-refreshes every 2s).</p>
<div class="metric live"><strong>Naive requests:</strong> %d &nbsp;|&nbsp; <strong>Last duration:</strong> %s</div>
<div class="metric live"><strong>Pooled requests:</strong> %d &nbsp;|&nbsp; <strong>Last duration:</strong> %s</div>
<div class="actions">
<a href="/naive?size=4096">Run Naive (4KB)</a>
<a href="/naive?size=65536">Run Naive (64KB)</a>
<a href="/pooled?size=4096">Run Pooled (4KB)</a>
<a href="/pooled?size=65536">Run Pooled (64KB)</a>
</div>

<h2>GC &amp; memory metrics</h2>
<div class="metric"><strong>Alloc (bytes):</strong> %d</div>
<div class="metric"><strong>TotalAlloc (bytes):</strong> %d</div>
<div class="metric"><strong>Sys (bytes):</strong> %d</div>
<div class="metric"><strong>NumGC:</strong> %d</div>
<div class="metric"><strong>GCCPUFraction:</strong> %f</div>
<div class="metric"><strong>PauseTotalNs:</strong> %d</div>
<div class="metric"><strong>HeapObjects:</strong> %d</div>
<div class="metric"><strong>LastGC:</strong> %s</div>
<p><em>Auto-refresh 2s.</em></p>
</body>
</html>`,
		naiveCnt, lastNaiveStr, pooledCnt, lastPooledStr,
		m.Alloc, m.TotalAlloc, m.Sys, m.NumGC, m.GCCPUFraction, m.PauseTotalNs, m.HeapObjects, lastGCStr)
	fmt.Fprint(w, html)
}
