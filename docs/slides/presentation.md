<!-- .slide: class="title-slide" -->
# Multi-Lang EDA SDK
## PoC for Func 2.0

<div class="author">Chris Suszyński</div>

---

# The Problem
## Did we learn from CloudEvents SDK?

<ul>
<li class="fragment">N languages = N implementations</li>
<li class="fragment">Features vary across SDKs</li>
<li class="fragment">Hard to keep up</li>
</ul>

---

<!-- .slide: data-background-color="#EE0000" -->
# The Solution
## Write core once, use across all SDKs

<ul>
<li class="fragment">Core in a single language</li>
<li class="fragment">Language SDKs just wrap the core</li>
<li class="fragment">Same behavior everywhere</li>
</ul>

---

# Architecture

<ol>
<li class="fragment"><strong>Core</strong> — shared logic in Rust</li>
<li class="fragment"><strong>SDKs</strong> — thin wrappers per language</li>
<li class="fragment"><strong>Integration</strong> — FFI or WASI</li>
</ol>

---

# What's Implemented
## Proof of concept, not production

<ul>
<li class="fragment">🟢 Routing — works end-to-end</li>
<li class="fragment">🟡 Config, Retry, Telemetry — stubs</li>
<li class="fragment">✅ Go FFI + Python FFI working</li>
</ul>

---

<!-- .slide: data-background-color="#151515" -->
# Proof Points
## Code sharing works

<ul>
<li class="fragment">Same routing logic → Go + Python</li>
<li class="fragment">Implement once → all benefit</li>
<li class="fragment">Architecture validated</li>
</ul>

---

# FFI vs WASI

<table>
<tr><th></th><th>FFI (Today)</th><th>WASI (Future)</th></tr>
<tr class="fragment"><td><strong>Status</strong></td><td>✅ Works now</td><td>⏳ Maturing</td></tr>
<tr class="fragment"><td><strong>Build</strong></td><td>Cumbersome, per-platform</td><td>Single WASM file</td></tr>
<tr class="fragment"><td><strong>Size</strong></td><td>~2MB each</td><td>~100KB total</td></tr>
<tr class="fragment"><td><strong>Hosts</strong></td><td>Any language</td><td>Rust only (P2)</td></tr>
<tr class="fragment"><td><strong>Debug</strong></td><td>C debuggers</td><td>Source maps (soon)</td></tr>
<tr class="fragment"><td><strong>Verdict</strong></td><td>Start here</td><td>Migrate later</td></tr>
</table>

---

# Developer Experience
## Business logic, not plumbing

```go
func Handle(event cloudevents.Event) error {
    // Your business logic here
}
```

<ul>
<li class="fragment">SDK handles Kafka, retries, routing</li>
<li class="fragment">YAML config for routing rules</li>
</ul>

---

<!-- .slide: class="demo-slide" data-background-color="#EE0000" -->
# 🎬 Live Demo

---

# What You Just Saw

<ul>
<li class="fragment">Go FFI consuming from Kafka</li>
<li class="fragment">Rust core making routing decisions</li>
<li class="fragment">Same logic shared with Python</li>
</ul>

---

# What's Next?

<ul>
<li class="fragment"><strong>Approve?</strong> Build it properly</li>
<li class="fragment">FFI SDKs for all target languages</li>
<li class="fragment">Migrate to WASI when Preview 3 lands</li>
</ul>

---

<!-- .slide: class="end-slide" data-background-color="#151515" -->
# Thanks!
## Questions?
