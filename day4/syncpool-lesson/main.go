package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// Live stats for dashboard
var (
	statsMu                sync.RWMutex
	buggyRequestCount      uint64
	fixedRequestCount      uint64
	lastBuggyDurationNs    int64
	lastFixedDurationNs    int64
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboardHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/buggy", buggyHandlerWithStats)
	mux.HandleFunc("/fixed", fixedHandlerWithStats)
	mux.HandleFunc("/debug/mem", handleMemStats)
	mux.HandleFunc("/api/stats", handleStatsJSON)
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func recordBuggyStats(d time.Duration) {
	statsMu.Lock()
	buggyRequestCount++
	lastBuggyDurationNs = d.Nanoseconds()
	statsMu.Unlock()
}

func recordFixedStats(d time.Duration) {
	statsMu.Lock()
	fixedRequestCount++
	lastFixedDurationNs = d.Nanoseconds()
	statsMu.Unlock()
}

func handleStatsJSON(w http.ResponseWriter, r *http.Request) {
	statsMu.RLock()
	defer statsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"buggy_request_count":     buggyRequestCount,
		"fixed_request_count":     fixedRequestCount,
		"last_buggy_duration_ns":  lastBuggyDurationNs,
		"last_fixed_duration_ns":  lastFixedDurationNs,
	})
}

func handleMemStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Alloc":         m.Alloc,
		"TotalAlloc":    m.TotalAlloc,
		"Sys":           m.Sys,
		"NumGC":         m.NumGC,
		"GCCPUFraction": m.GCCPUFraction,
		"PauseTotalNs":  m.PauseTotalNs,
		"HeapObjects":   m.HeapObjects,
		"LastGC":        time.Unix(0, int64(m.LastGC)).Format(time.RFC3339Nano),
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	lastGCStr := time.Unix(0, int64(m.LastGC)).Format(time.RFC3339Nano)
	statsMu.RLock()
	buggyCnt := buggyRequestCount
	fixedCnt := fixedRequestCount
	lastBuggyNs := lastBuggyDurationNs
	lastFixedNs := lastFixedDurationNs
	statsMu.RUnlock()
	lastBuggyStr := "—"
	if lastBuggyNs > 0 {
		lastBuggyStr = time.Duration(lastBuggyNs).String()
	}
	lastFixedStr := "—"
	if lastFixedNs > 0 {
		lastFixedStr = time.Duration(lastFixedNs).String()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>sync.Pool Lesson Dashboard</title>
<meta http-equiv="refresh" content="2">
<style>
body{font-family:system-ui,sans-serif;margin:24px;background:#f8f9fa;color:#212529;}
h1{color:#212529;}
h2{color:#495057;font-size:1.1rem;margin-top:24px;}
.info{background:#fff;border:1px solid #dee2e6;border-radius:8px;padding:16px;margin:12px 0;}
.metric{background:#fff;border:1px solid #dee2e6;padding:12px 16px;margin:8px 0;border-radius:8px;}
.nav{margin:16px 0;}
.nav a{margin-right:16px;}
.live{background:#fff3cd;border:1px solid #ffc107;}
.actions a{display:inline-block;margin-right:8px;padding:8px 16px;background:#212529;color:#fff;border-radius:6px;text-decoration:none;}
</style>
</head>
<body>
<h1>sync.Pool Lesson Dashboard</h1>
<div class="info"><strong>About:</strong> Buggy (no buffer reset) vs Fixed (reset before Put). Use actions below to generate load; dashboard updates every 2s.</div>
<p class="nav"><a href="/">Dashboard</a> | <a href="/debug/mem">JSON MemStats</a> | <a href="/api/stats">JSON Live Stats</a></p>
<h2>Live request metrics</h2>
<div class="metric live"><strong>Buggy requests:</strong> %d | <strong>Last duration:</strong> %s</div>
<div class="metric live"><strong>Fixed requests:</strong> %d | <strong>Last duration:</strong> %s</div>
<div class="actions">
<a href="/buggy" target="_blank">Run Buggy (GET)</a>
<a href="/fixed" target="_blank">Run Fixed (GET)</a>
</div>
<h2>GC &amp; memory metrics</h2>
<div class="metric"><strong>Alloc:</strong> %d</div>
<div class="metric"><strong>TotalAlloc:</strong> %d</div>
<div class="metric"><strong>NumGC:</strong> %d</div>
<div class="metric"><strong>PauseTotalNs:</strong> %d</div>
<div class="metric"><strong>HeapObjects:</strong> %d</div>
<div class="metric"><strong>LastGC:</strong> %s</div>
</body>
</html>`,
		buggyCnt, lastBuggyStr, fixedCnt, lastFixedStr,
		m.Alloc, m.TotalAlloc, m.NumGC, m.PauseTotalNs, m.HeapObjects, lastGCStr)
	fmt.Fprint(w, html)
}

func buggyHandlerWithStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { recordBuggyStats(time.Since(start)) }()
	buggyHandler(w, r)
}

func fixedHandlerWithStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() { recordFixedStats(time.Since(start)) }()
	fixedHandler(w, r)
}

func buggyHandler(w http.ResponseWriter, r *http.Request) {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() { bufferPool.Put(buf) }()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	buf.Write(bodyBytes)
	requestID := fmt.Sprintf("REQ-%d", time.Now().UnixNano()%100000)
	buf.WriteString(fmt.Sprintf(" [Processed by %s]", requestID))
	response := fmt.Sprintf("Buggy response (len: %d): %s", buf.Len(), buf.String())
	fmt.Fprint(w, response)
}

func fixedHandler(w http.ResponseWriter, r *http.Request) {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() { buf.Reset(); bufferPool.Put(buf) }()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	buf.Write(bodyBytes)
	requestID := fmt.Sprintf("REQ-%d", time.Now().UnixNano()%100000)
	buf.WriteString(fmt.Sprintf(" [Processed by %s]", requestID))
	response := fmt.Sprintf("Fixed response (len: %d): %s", buf.Len(), buf.String())
	fmt.Fprint(w, response)
}
