package main

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"
)

// User represents a simple user struct.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PopulateUserReflect populates a User struct from a map using reflection.
func PopulateUserReflect(data map[string]interface{}) (*User, error) {
	user := &User{}
	val := reflect.ValueOf(user).Elem()

	for key, mapVal := range data {
		field := val.FieldByName(key)
		if !field.IsValid() {
			continue
		}
		if !field.CanSet() {
			return nil, fmt.Errorf("field %s cannot be set", key)
		}
		mapValReflect := reflect.ValueOf(mapVal)
		if mapValReflect.Type().AssignableTo(field.Type()) {
			field.Set(mapValReflect)
		} else if mapValReflect.Type().ConvertibleTo(field.Type()) {
			field.Set(mapValReflect.Convert(field.Type()))
		} else {
			return nil, fmt.Errorf("cannot assign map value of type %s to field %s of type %s", mapValReflect.Type(), key, field.Type())
		}
	}
	return user, nil
}

// PopulateUserDirect populates a User struct from a map using direct assignment.
func PopulateUserDirect(data map[string]interface{}) (*User, error) {
	user := &User{}
	if id, ok := data["ID"].(float64); ok {
		user.ID = int(id)
	} else if id, ok := data["ID"].(int); ok {
		user.ID = id
	} else if _, ok := data["ID"]; ok {
		return nil, errors.New("ID is not an integer or float64")
	}
	if name, ok := data["Name"].(string); ok {
		user.Name = name
	} else if _, ok := data["Name"]; ok {
		return nil, errors.New("Name is not a string")
	}
	if email, ok := data["Email"].(string); ok {
		user.Email = email
	} else if _, ok := data["Email"]; ok {
		return nil, errors.New("Email is not a string")
	}
	return user, nil
}

const demoIterations = 50000
const backgroundBatchSize = 10000

var backgroundInterval = 3 * time.Second

var (
	statsMu              sync.RWMutex
	reflectCount         int64
	directCount          int64
	lastReflectOps       int64
	lastReflectNs        int64
	lastDirectOps        int64
	lastDirectNs         int64
	backgroundRoundCount int64
	lastBackgroundRun    time.Time
)

var testMap = map[string]interface{}{
	"ID":    float64(123),
	"Name":  "Alice Wonderland",
	"Email": "alice@example.com",
}

func main() {
	go runBackgroundComparison()
	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboardHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/demo/reflect", demoReflectHandler)
	mux.HandleFunc("/demo/direct", demoDirectHandler)
	mux.HandleFunc("/api/stats", statsAPIHandler)
	port := ":8080"
	fmt.Printf("Day 8 Reflect vs Direct server on %s\n", port)
	http.ListenAndServe(port, mux)
}

// runBackgroundComparison runs reflect and direct batches every backgroundInterval.
// Metrics update automatically so users see a continuous comparison without clicking.
func runBackgroundComparison() {
	ticker := time.NewTicker(backgroundInterval)
	defer ticker.Stop()
	for range ticker.C {
		opsR, nsR := runReflectBatch(backgroundBatchSize)
		opsD, nsD := runDirectBatch(backgroundBatchSize)
		statsMu.Lock()
		reflectCount++
		directCount++
		lastReflectOps = opsR
		lastReflectNs = nsR
		lastDirectOps = opsD
		lastDirectNs = nsD
		backgroundRoundCount++
		lastBackgroundRun = time.Now()
		statsMu.Unlock()
	}
}

func runReflectBatch(n int) (ops int64, elapsedNs int64) {
	start := time.Now()
	for i := 0; i < n; i++ {
		_, _ = PopulateUserReflect(testMap)
	}
	return int64(n), time.Since(start).Nanoseconds()
}

func runDirectBatch(n int) (ops int64, elapsedNs int64) {
	start := time.Now()
	for i := 0; i < n; i++ {
		_, _ = PopulateUserDirect(testMap)
	}
	return int64(n), time.Since(start).Nanoseconds()
}

func runReflectDemo() (ops int64, elapsedNs int64) { return runReflectBatch(demoIterations) }
func runDirectDemo() (ops int64, elapsedNs int64) { return runDirectBatch(demoIterations) }

func demoReflectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ops, elapsedNs := runReflectDemo()
	statsMu.Lock()
	reflectCount++
	lastReflectOps = ops
	lastReflectNs = elapsedNs
	statsMu.Unlock()
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/dashboard#last-run", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ops":%d,"duration_ns":%d,"ns_per_op":%.2f}`, ops, elapsedNs, float64(elapsedNs)/float64(ops))
}

func demoDirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ops, elapsedNs := runDirectDemo()
	statsMu.Lock()
	directCount++
	lastDirectOps = ops
	lastDirectNs = elapsedNs
	statsMu.Unlock()
	if r.Method == http.MethodGet {
		http.Redirect(w, r, "/dashboard#last-run", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"ops":%d,"duration_ns":%d,"ns_per_op":%.2f}`, ops, elapsedNs, float64(elapsedNs)/float64(ops))
}

func statsAPIHandler(w http.ResponseWriter, r *http.Request) {
	statsMu.RLock()
	defer statsMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"comparison_rounds":%d,"last_background_run":%q,"last_reflect_ops":%d,"last_reflect_ns":%d,"last_direct_ops":%d,"last_direct_ns":%d}`,
		backgroundRoundCount, lastBackgroundRun.Format(time.RFC3339), lastReflectOps, lastReflectNs, lastDirectOps, lastDirectNs)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	statsMu.RLock()
	rounds := backgroundRoundCount
	lastRun := lastBackgroundRun
	lro := lastReflectOps
	lrn := lastReflectNs
	ldo := lastDirectOps
	ldn := lastDirectNs
	statsMu.RUnlock()

	reflectNsPerOp := "—"
	if lro > 0 && lrn > 0 {
		reflectNsPerOp = fmt.Sprintf("%.2f", float64(lrn)/float64(lro))
	}
	directNsPerOp := "—"
	if ldo > 0 && ldn > 0 {
		directNsPerOp = fmt.Sprintf("%.2f", float64(ldn)/float64(ldo))
	}
	lastRunStr := "—"
	if !lastRun.IsZero() {
		lastRunStr = lastRun.Format("15:04:05")
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Reflect vs Direct Dashboard</title>
<meta http-equiv="refresh" content="5">
<meta charset="utf-8">
<style>
body{font-family:system-ui,sans-serif;margin:0;padding:24px;background:#f8f9fa;color:#212529;}
.dashboard-container{max-width:1200px;margin:0 auto;}
h1{color:#212529;} h2{color:#495057;font-size:1.1rem;margin-top:24px;}
.about{background:#fff;border:1px solid #dee2e6;border-radius:8px;padding:16px;margin:12px 0;line-height:1.6;}
.about strong{display:block;margin-bottom:6px;}
.how-metrics{background:#e8f4fd;border:1px solid #0d6efd;border-radius:8px;padding:12px 16px;margin:12px 0;font-size:0.95rem;line-height:1.5;}
.metric{background:#fff;border:1px solid #dee2e6;padding:12px 16px;margin:8px 0;border-radius:8px;}
.metric.reflect{border-left:4px solid #dc3545;}
.metric.direct{border-left:4px solid #198754;}
.actions{margin:16px 0;}
.actions a,.actions button{display:inline-block;margin-right:8px;padding:8px 16px;border-radius:6px;font-size:0.9rem;cursor:pointer;font-family:inherit;border:none;}
.actions a{background:#495057;color:#fff;text-decoration:none;}
.actions button.primary{background:#212529;color:#fff;}
.actions button.primary:hover{background:#343a40;}
.nav{margin:16px 0;}
.workflow-section{margin-top:32px;padding-top:24px;border-top:1px solid #dee2e6;}
.workflow-title{font-size:1.1rem;font-weight:600;margin-bottom:12px;color:#212529;}
.workflow-info{background:#f8f9fa;border:1px solid #dee2e6;border-radius:8px;padding:12px 16px;margin:12px 0;font-size:0.9rem;line-height:1.5;}
.workflow-diagram{display:flex;flex-wrap:nowrap;align-items:stretch;gap:6px;margin:16px 0;padding:14px;background:#fff;border:1px solid #dee2e6;border-radius:8px;overflow-x:auto;overflow-y:hidden;}
.workflow-diagram::-webkit-scrollbar{height:8px;}
.workflow-diagram::-webkit-scrollbar-thumb{background:#adb5bd;border-radius:4px;}
.workflow-step{flex-shrink:0;min-width:88px;padding:10px 12px;background:linear-gradient(180deg,#f0f4ff 0%%,#e2e8f0 100%%);border:1px solid #94a3b8;border-radius:8px;text-align:center;box-shadow:0 1px 3px rgba(0,0,0,0.08);}
.workflow-step .step-num{display:block;font-size:0.65rem;font-weight:700;color:#475569;letter-spacing:0.03em;margin-bottom:2px;}
.workflow-step .step-title{display:block;font-weight:600;color:#1e293b;font-size:0.8rem;}
.workflow-step .step-desc{display:block;font-size:0.65rem;color:#64748b;margin-top:2px;}
.workflow-arrow{flex-shrink:0;color:#94a3b8;font-weight:700;font-size:1.1rem;display:flex;align-items:center;}
.workflow-goal{background:linear-gradient(180deg,#dcfce7 0%%,#bbf7d0 100%%);border-color:#22c55e;}
.workflow-goal .step-num{color:#166534;}
@keyframes wfFade{0%%{opacity:0;transform:translateY(8px) scale(0.96);}100%%{opacity:1;transform:translateY(0) scale(1);}}
@keyframes wfPulse{0%%{box-shadow:0 1px 3px rgba(0,0,0,0.08);}50%%{box-shadow:0 0 0 4px rgba(59,130,246,0.35);}100%%{box-shadow:0 1px 3px rgba(0,0,0,0.08);}}
.wf{opacity:0;animation-fill-mode:forwards;}
.wf.d0{animation:wfFade 0.7s ease-out 0s forwards,wfPulse 3s ease-in-out 0.8s infinite;}
.wf.d1{animation:wfFade 0.7s ease-out 0.1s forwards,wfPulse 3s ease-in-out 0.9s infinite;}
.wf.d2{animation:wfFade 0.7s ease-out 0.2s forwards,wfPulse 3s ease-in-out 1s infinite;}
.wf.d3{animation:wfFade 0.7s ease-out 0.3s forwards,wfPulse 3s ease-in-out 1.1s infinite;}
.wf.d4{animation:wfFade 0.7s ease-out 0.4s forwards,wfPulse 3s ease-in-out 1.2s infinite;}
.wf.d5{animation:wfFade 0.7s ease-out 0.5s forwards,wfPulse 3s ease-in-out 1.3s infinite;}
.wf.d6{animation:wfFade 0.7s ease-out 0.6s forwards,wfPulse 3s ease-in-out 1.4s infinite;}
.workflow-arrow{animation:wfPulse 2.5s ease-in-out 0.5s infinite;}
</style>
</head>
<body>
<div class="dashboard-container">
<h1>Reflect vs Direct Dashboard</h1>

<div class="about">
<strong>What is this application?</strong>
This app compares two ways to turn key-value data (e.g. from JSON or a map) into a typed Go struct: <strong>Reflect</strong> (using the <code>reflect</code> package to inspect and set fields by name at runtime) and <strong>Direct</strong> (explicit type assertions and assignment). Both produce the same result; the difference is performance and when each approach is appropriate.
<strong>Why does it matter?</strong>
Reflection is flexible and works for any struct without code changes, but it is slower and allocates more memory. Direct code is faster and allocation-friendly but is written for one struct. In high-throughput or latency-sensitive paths you want direct; for config loading or generic tools, reflect is often acceptable.
<strong>How does it work?</strong>
The server runs a <em>continuous background comparison</em>: every 3 seconds it executes a batch of reflect-based and direct-based population, then updates the metrics below. You see a live performance comparison without clicking anything. The numbers show how many nanoseconds each approach takes per operation (ns/op)—lower is better.
</div>

<div class="how-metrics">
<strong>How the metrics work</strong><br>
The values below are updated automatically by a background process that runs every 3 seconds. It runs both approaches (reflect and direct) on the same data and records the time taken. You are not filling a placeholder by clicking a button—you are watching a real-time comparison. Optional: use the buttons under the metrics to run a heavier on-demand comparison (50k ops each) if you want to stress the server.
</div>

<h2>Live comparison metrics</h2>
<div id="metric-rounds" class="metric"><strong>Comparison rounds:</strong> %d (every 3s) | <strong>Last run:</strong> %s</div>
<div id="metric-reflect" class="metric reflect"><strong>Reflect path</strong> — Last batch: %d ops in %d ns → <strong>ns/op: %s</strong></div>
<div id="metric-direct" class="metric direct"><strong>Direct path</strong> — Last batch: %d ops in %d ns → <strong>ns/op: %s</strong></div>
<div class="actions">
<strong>Optional on-demand run:</strong>
<button type="button" class="primary demo-btn" data-url="/demo/reflect" data-label="Run heavy Reflect (50k ops)">Run heavy Reflect (50k ops)</button>
<button type="button" class="primary demo-btn" data-url="/demo/direct" data-label="Run heavy Direct (50k ops)">Run heavy Direct (50k ops)</button>
</div>
<script>
function formatLastRun(s){
  if(!s || s.indexOf('0001-01-01')===0) return '—';
  var d=new Date(s);
  return isNaN(d.getTime()) ? '—' : d.toTimeString().slice(0,8);
}
function updateMetrics(){
  fetch('/api/stats').then(function(r){ return r.json(); }).then(function(d){
    var roundsEl=document.getElementById('metric-rounds');
    var reflectEl=document.getElementById('metric-reflect');
    var directEl=document.getElementById('metric-direct');
    if(roundsEl) roundsEl.innerHTML='<strong>Comparison rounds:</strong> '+d.comparison_rounds+' (every 3s) | <strong>Last run:</strong> '+formatLastRun(d.last_background_run);
    if(reflectEl) reflectEl.innerHTML='<strong>Reflect path</strong> — Last batch: '+d.last_reflect_ops+' ops in '+d.last_reflect_ns+' ns → <strong>ns/op: '+(d.last_reflect_ops>0 ? (d.last_reflect_ns/d.last_reflect_ops).toFixed(2) : '—')+'</strong>';
    if(directEl) directEl.innerHTML='<strong>Direct path</strong> — Last batch: '+d.last_direct_ops+' ops in '+d.last_direct_ns+' ns → <strong>ns/op: '+(d.last_direct_ops>0 ? (d.last_direct_ns/d.last_direct_ops).toFixed(2) : '—')+'</strong>';
  }).catch(function(){});
}
document.querySelectorAll('.demo-btn').forEach(function(btn){
  btn.addEventListener('click',function(e){
    e.preventDefault();
    var url=this.getAttribute('data-url');
    var label=this.getAttribute('data-label');
    if(!url) return;
    this.textContent='Running...';
    this.disabled=true;
    var self=this;
    fetch(url).then(function(){ updateMetrics(); }).then(function(){ self.textContent=label; self.disabled=false; }).catch(function(){ self.textContent=label; self.disabled=false; });
  });
});
</script>
<p class="nav"><a href="/api/stats">JSON stats</a></p>

<div class="workflow-section">
<div class="workflow-title">Application workflow</div>
<div class="workflow-info">This diagram describes the <em>application</em> flow—how data moves and how the two approaches are compared—not the code structure.</div>
<div class="workflow-diagram">
  <div class="workflow-step wf d0"><span class="step-num">1</span><span class="step-title">Input</span><span class="step-desc">Key-value data (e.g. API/JSON)</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step wf d1"><span class="step-num">2</span><span class="step-title">Goal</span><span class="step-desc">Populate typed struct</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step wf d2"><span class="step-num">3a</span><span class="step-title">Reflect path</span><span class="step-desc">Generic, any struct; slower</span></div>
  <div class="workflow-arrow">+</div>
  <div class="workflow-step wf d3"><span class="step-num">3b</span><span class="step-title">Direct path</span><span class="step-desc">Explicit, this struct; fast</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step wf d4"><span class="step-num">4</span><span class="step-title">Compare</span><span class="step-desc">Same batch size, measure time</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step wf d5"><span class="step-num">5</span><span class="step-title">Report</span><span class="step-desc">ns/op on dashboard</span></div>
  <div class="workflow-arrow">→</div>
  <div class="workflow-step workflow-goal wf d6"><span class="step-num">Takeaway</span><span class="step-title">Use direct in hot paths</span><span class="step-desc">Use reflect for config/generic</span></div>
</div>
<div class="workflow-info"><strong>In short:</strong> Data arrives → we need a struct. Two ways: reflect (flexible, slower) and direct (fast, struct-specific). This app runs both continuously and shows you the performance gap so you can choose the right tool for production.</div>
</div>

</div>
</body>
</html>`,
		rounds, lastRunStr,
		lro, lrn, reflectNsPerOp,
		ldo, ldn, directNsPerOp)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
