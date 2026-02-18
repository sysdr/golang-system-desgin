package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"day6_memory_alignment/struct_analyzer"
)

var (
	statsMu         sync.RWMutex
	lastResults     []struct_analyzer.StructResult
	analysisCount   uint64
	lastAnalysisAt  time.Time
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboardHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/api/run-analysis", runAnalysisHandler)
	mux.HandleFunc("/api/stats", statsJSONHandler)
	mux.HandleFunc("/api/results", resultsJSONHandler)
	mux.HandleFunc("/debug/mem", memStatsHandler)
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func runAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	results := struct_analyzer.RunAnalysisToResults()
	statsMu.Lock()
	lastResults = results
	analysisCount++
	lastAnalysisAt = time.Now()
	statsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": true, "count": len(results), "analysis_count": analysisCount,
	})
}

func statsJSONHandler(w http.ResponseWriter, r *http.Request) {
	statsMu.RLock()
	defer statsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"analysis_count":   analysisCount,
		"last_analysis_at":  lastAnalysisAt.Format(time.RFC3339),
		"structs_analyzed": len(lastResults),
	})
}

func resultsJSONHandler(w http.ResponseWriter, r *http.Request) {
	statsMu.RLock()
	defer statsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	type simpleResult struct {
		Name string `json:"name"`
		Size uintptr `json:"size"`
		Alignment uintptr `json:"alignment"`
	}
	out := make([]simpleResult, len(lastResults))
	for i := range lastResults {
		out[i] = simpleResult{
			Name: lastResults[i].Name,
			Size: lastResults[i].Size,
			Alignment: lastResults[i].Alignment,
		}
	}
	json.NewEncoder(w).Encode(out)
}

func memStatsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Alloc": m.Alloc, "TotalAlloc": m.TotalAlloc, "Sys": m.Sys,
		"NumGC": m.NumGC, "HeapObjects": m.HeapObjects,
	})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	statsMu.RLock()
	results := lastResults
	count := analysisCount
	lastAt := lastAnalysisAt
	statsMu.RUnlock()

	lastAtStr := "Never"
	if !lastAt.IsZero() {
		lastAtStr = lastAt.Format(time.RFC3339)
	}

	// Build struct metrics HTML
	structRows := ""
	if len(results) > 0 {
		for _, r := range results {
			structRows += fmt.Sprintf(
				`<div class="metric"><strong>%s</strong> Size: %d bytes, Alignment: %d</div>`,
				r.Name, r.Size, r.Alignment,
			)
		}
	} else {
		structRows = `<div class="metric live">No analysis yet. Click "Run Analysis" below to update metrics.</div>`
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Memory Alignment Dashboard</title>
<meta http-equiv="refresh" content="10">
<style>
body{font-family:system-ui,sans-serif;margin:24px;background:#f8f9fa;color:#212529;}
h1{color:#212529;} h2{color:#495057;font-size:1.1rem;margin-top:24px;}
.info{background:#fff;border:1px solid #dee2e6;border-radius:8px;padding:16px;margin:12px 0;}
.metric{background:#fff;border:1px solid #dee2e6;padding:12px 16px;margin:8px 0;border-radius:8px;}
.nav{margin:16px 0;} .nav a{margin-right:16px;}
.live{background:#fff3cd;border:1px solid #ffc107;}
.actions a{display:inline-block;margin-right:8px;padding:8px 16px;background:#212529;color:#fff;border-radius:6px;text-decoration:none;}
.workflow-info{background:#e8f4fd;border:1px solid #0d6efd;border-radius:8px;padding:12px 16px;margin:12px 0;font-size:0.95rem;line-height:1.5;}
.workflow-info code{background:#fff;padding:2px 6px;border-radius:4px;}
.workflow-diagram{display:flex;flex-wrap:wrap;align-items:center;gap:4px;margin:16px 0;padding:16px;background:#fff;border:1px solid #dee2e6;border-radius:8px;overflow-x:auto;}
.workflow-step{min-width:100px;padding:10px 12px;background:linear-gradient(180deg,#f0f4ff 0%%,#e2e8f0 100%%);border:1px solid #94a3b8;border-radius:6px;text-align:center;box-shadow:0 1px 3px rgba(0,0,0,0.08);}
.workflow-step .step-num{display:block;font-size:0.7rem;font-weight:700;color:#475569;letter-spacing:0.05em;margin-bottom:4px;}
.workflow-step .step-title{display:block;font-weight:600;color:#1e293b;font-size:0.9rem;}
.workflow-step .step-desc{display:block;font-size:0.75rem;color:#64748b;margin-top:4px;}
.workflow-goal{background:linear-gradient(180deg,#dcfce7 0%%,#bbf7d0 100%%);border-color:#22c55e;}
.workflow-goal .step-num{color:#166534;}
.workflow-arrow{color:#94a3b8;font-weight:700;font-size:1.1rem;}
@keyframes workflowFade{0%%{opacity:0;transform:translateY(8px);}100%%{opacity:1;transform:translateY(0);}}
@keyframes workflowPulse{0%%{box-shadow:0 1px 3px rgba(0,0,0,0.08);}50%%{box-shadow:0 0 0 2px rgba(15,23,42,0.12);}100%%{box-shadow:0 1px 3px rgba(0,0,0,0.08);}}
.workflow-anim{opacity:0;}
.workflow-anim.workflow-delay-0{animation:workflowFade 0.5s ease-out 0s forwards,workflowPulse 10s ease-in-out 1.5s infinite;}
.workflow-anim.workflow-delay-1{animation:workflowFade 0.5s ease-out 0.2s forwards,workflowPulse 10s ease-in-out 1.5s infinite;}
.workflow-anim.workflow-delay-2{animation:workflowFade 0.5s ease-out 0.4s forwards,workflowPulse 10s ease-in-out 1.5s infinite;}
.workflow-anim.workflow-delay-3{animation:workflowFade 0.5s ease-out 0.6s forwards,workflowPulse 10s ease-in-out 1.5s infinite;}
.workflow-anim.workflow-delay-4{animation:workflowFade 0.5s ease-out 0.8s forwards,workflowPulse 10s ease-in-out 1.5s infinite;}
.workflow-anim.workflow-delay-5{animation:workflowFade 0.5s ease-out 1s forwards,workflowPulse 10s ease-in-out 1.5s infinite;}
.workflow-anim.workflow-delay-6{animation:workflowFade 0.5s ease-out 1.2s forwards,workflowPulse 10s ease-in-out 1.5s infinite;}
.workflow-arrow{animation:workflowPulse 10s ease-in-out 2s infinite;}
.workflow-why{background:#fefce8;border:1px solid #eab308;border-radius:8px;padding:12px 16px;margin-top:12px;font-size:0.9rem;line-height:1.5;color:#713f12;}
</style>
</head>
<body>
<h1>Memory Alignment Dashboard</h1>
<div class="info"><strong>About:</strong> Struct layout and padding. Run analysis to see sizes; dashboard updates every 10s.</div>
<p class="nav"><a href="/">Dashboard</a> | <a href="/debug/mem">JSON MemStats</a> | <a href="/api/stats">JSON Stats</a> | <a href="/api/results">JSON Results</a></p>
<h2>Live struct metrics</h2>
%s
<div class="metric"><strong>Analyses run:</strong> %d | <strong>Last run:</strong> %s</div>
<div class="actions">
<a href="/api/run-analysis" target="_blank">Run Analysis</a>
</div>
<h2>GC &amp; memory</h2>
<div class="metric"><strong>Alloc:</strong> %d | <strong>TotalAlloc:</strong> %d | <strong>NumGC:</strong> %d | <strong>HeapObjects:</strong> %d</div>

<h2>Project workflow</h2>
<div class="workflow-info">This app follows a <strong>Data Layout Optimization Pipeline</strong>: unoptimized structs are analyzed with <code>reflect</code> and <code>unsafe</code>, memory map is displayed, then field reordering minimizes padding so the CPU gets aligned access and fewer cache-line splits.</div>
<div class="workflow-diagram">
  <div class="workflow-step workflow-anim workflow-delay-0"><span class="step-num">START</span><span class="step-title">Begin</span><span class="step-desc">Pipeline entry</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step workflow-anim workflow-delay-1"><span class="step-num">1</span><span class="step-title">Define struct</span><span class="step-desc">Unoptimized field order</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step workflow-anim workflow-delay-2"><span class="step-num">2</span><span class="step-title">Reflective analysis</span><span class="step-desc">Size & padding bytes</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step workflow-anim workflow-delay-3"><span class="step-num">3</span><span class="step-title">Memory map</span><span class="step-desc">Visualize padding</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step workflow-anim workflow-delay-4"><span class="step-num">4</span><span class="step-title">Reorder fields</span><span class="step-desc">Descending size</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step workflow-anim workflow-delay-5"><span class="step-num">5</span><span class="step-title">Benchmark</span><span class="step-desc">Verify footprint</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step workflow-goal workflow-anim workflow-delay-6"><span class="step-num">GOAL</span><span class="step-title">0-byte padding</span><span class="step-desc">Aligned, cache-friendly</span></div>
</div>
<div class="workflow-why">Why: Aligned data fits in a single cache line (single-cycle fetch); unaligned data straddles 64-byte boundaries and costs 2+ bus transactions and CPU stalls. Reorder & pad to avoid false sharing and maximize throughput.</div>
</body>
</html>`,
		structRows, count, lastAtStr,
		m.Alloc, m.TotalAlloc, m.NumGC, m.HeapObjects,
	)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
