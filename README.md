# FuzzMaster

FuzzMaster is a fast, multi‑threaded web fuzzing tool written in Go.  
It is designed for bug bounty and security testing workflows where you need visibility, control, and clean output.

It supports subdomain‑based fuzzing, directory and extension discovery, soft‑404 detection, rate limiting, and optional Nuclei integration.

<img width="1097" height="546" alt="image" src="https://github.com/user-attachments/assets/9d49bc28-3ce5-466f-9a04-2c5635de4a09" />


---

## ✨ Features

- Subdomain‑based directory fuzzing
- Concurrent workers with rate limiting
- Soft‑404 detection per host
- Colorful, readable terminal output
- Status code and response size visibility
- Live request counters (2xx / 3xx / 4xx / 5xx)
- JSON output for reporting
- Safe Ctrl+C handling (graceful shutdown)
- Optional Nuclei execution on valid hits

---

## 📦 Installation

Make sure Go is installed (Go 1.20+ recommended).

```bash
git clone https://github.com/yourusername/fuzzmaster.git
cd fuzzmaster
go build -o fuzzmaster
