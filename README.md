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
go build -o fuzzmaster.exe (for windows)


On Windows:

go build -o fuzzmaster.exe

🚀 Usage
./fuzzmaster -subs subs.txt -dirs dirs.txt

Example (Windows)
.\fuzzmaster.exe `
  -subs D:\SUBDOMAINS\apple.com.txt `
  -dirs C:\wordlists\dirs.txt `
  -rate 10 `
  -workers 20

🧩 Flags
Flag	Description
-subs	File containing subdomains
-dirs	Directory wordlist
-ext	Extensions to try (default: php,bak,zip)
-params	Query parameters to fuzz
-rate	Requests per second
-workers	Number of concurrent workers
-json	JSON output file (default: out.json)
-nuclei	Run nuclei on valid hits
🖥 Output
Alive Check
[ALIVE] example.com → 200 (1245 bytes)

Requests
[REQ] 403 https://example.com/admin (0 bytes)
[HIT] https://example.com/actuator

Summary
========== SUMMARY ==========
Total Requests : 1342
200s           : 17
300s           : 41
400s           : 1201
500s           : 83
=============================

📄 JSON Output

Results are saved in JSON format for easy reporting:

{
  "url": "https://example.com/actuator",
  "status": 200,
  "length": 532,
  "soft_404": false
}

⚠️ Disclaimer

This tool is intended for authorized security testing only.
Do not use it against systems you do not own or have explicit permission to test.
