package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

/* ================= COLORS ================= */

const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	bold    = "\033[1m"
)

func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return yellow
	case code >= 400 && code < 500:
		return magenta
	default:
		return red
	}
}

/* ================= STRUCTS ================= */

type Job struct {
	URL     string
	Soft404 int
}

type Result struct {
	URL     string `json:"url"`
	Status  int    `json:"status"`
	Length  int    `json:"length"`
	Soft404 bool   `json:"soft_404"`
}

/* ================= FLAGS ================= */

var (
	subsFile  = flag.String("subs", "", "subdomains file")
	dirsFile  = flag.String("dirs", "", "directories file")
	exts      = flag.String("ext", "php,bak,zip", "extensions")
	params    = flag.String("params", "", "params")
	rate      = flag.Int("rate", 5, "req/sec")
	jsonOut   = flag.String("json", "out.json", "json output")
	runNuclei = flag.Bool("nuclei", false, "run nuclei")
	workers   = flag.Int("workers", 10, "workers")
)

/* ================= GLOBALS ================= */

var (
	client  *http.Client
	results []Result
	rMu     sync.Mutex
	wg      sync.WaitGroup

	totalReq, c2xx, c3xx, c4xx, c5xx int
	cMu sync.Mutex
)

/* ================= HELPERS ================= */

func initClient() {
	client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func load(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(red+"[ERR]"+reset, "cannot open:", path)
		os.Exit(1)
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if s := strings.TrimSpace(sc.Text()); s != "" {
			out = append(out, s)
		}
	}
	return out
}

/* ================= ALIVE CHECK ================= */

func isAlive(host string) bool {
	url := "https://" + host
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(red+"[DEAD]"+reset, host)
		return false
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	length := int(resp.ContentLength)

	fmt.Printf(
		green+"[ALIVE]"+reset+" %s → %s%d%s %s(%d bytes)%s\n",
		host,
		statusColor(code), code, reset,
		blue, length, reset,
	)
	return true
}

func soft404Code(base string) int {
	testURL := fmt.Sprintf("%s/%d-notfound", base, time.Now().UnixNano())
	resp, err := client.Get(testURL)
	if err != nil {
		return -1
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

/* ================= WORKER ================= */

func worker(ctx context.Context, jobs <-chan Job, limiter *time.Ticker) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case job, ok := <-jobs:
			if !ok {
				return
			}

			<-limiter.C

			resp, err := client.Get(job.URL)
			if err != nil {
				fmt.Println(red+"[ERR]"+reset, job.URL)
				continue
			}

			code := resp.StatusCode
			length := int(resp.ContentLength)
			resp.Body.Close()

			cMu.Lock()
			totalReq++
			switch {
			case code >= 200 && code < 300:
				c2xx++
			case code >= 300 && code < 400:
				c3xx++
			case code >= 400 && code < 500:
				c4xx++
			default:
				c5xx++
			}
			cMu.Unlock()

			fmt.Printf(
				cyan+"[REQ]"+reset+" %s%d%s %s %s(%d bytes)%s\n",
				statusColor(code), code, reset,
				job.URL,
				blue, length, reset,
			)

			if code == job.Soft404 {
				continue
			}

			if code >= 200 && code < 400 {
				fmt.Println(bold+green+"[HIT]"+reset, job.URL)

				rMu.Lock()
				results = append(results, Result{
					URL:     job.URL,
					Status:  code,
					Length:  length,
					Soft404: false,
				})
				rMu.Unlock()

				if *runNuclei {
					_ = exec.Command("nuclei", "-u", job.URL).Start()
				}
			}
		}
	}
}

/* ================= OUTPUT ================= */

func saveJSON() {
	f, err := os.Create(*jsonOut)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(results)
	fmt.Println(green+"[INFO]"+reset, "saved", *jsonOut)
}

func printSummary() {
	fmt.Println(bold + "\n========== SUMMARY ==========" + reset)
	fmt.Printf("Total Requests : %d\n", totalReq)
	fmt.Printf("200s           : %d\n", c2xx)
	fmt.Printf("300s           : %d\n", c3xx)
	fmt.Printf("400s           : %d\n", c4xx)
	fmt.Printf("500s           : %d\n", c5xx)
	fmt.Println(bold + "=============================" + reset)
}

/* ================= MAIN ================= */

func main() {
	flag.Parse()
	initClient()

	subs := load(*subsFile)
	dirs := load(*dirsFile)
	extList := strings.Split(*exts, ",")
	paramList := strings.Split(*params, ",")

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	go func() {
		<-sig
		fmt.Println(yellow + "\n[!] Ctrl+C detected, shutting down..." + reset)
		cancel()
	}()

	jobs := make(chan Job, 1000)
	limiter := time.NewTicker(time.Second / time.Duration(*rate))
	defer limiter.Stop()

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(ctx, jobs, limiter)
	}

	for _, s := range subs {
		if !isAlive(s) {
			continue
		}

		base := "https://" + s
		soft := soft404Code(base)

		for _, d := range dirs {
			jobs <- Job{base + "/" + d, soft}

			for _, e := range extList {
				jobs <- Job{base + "/" + d + "." + e, soft}
			}

			for _, p := range paramList {
				if p != "" {
					jobs <- Job{base + "/" + d + "?" + p + "=test", soft}
				}
			}
		}
	}

	close(jobs)
	wg.Wait()
	saveJSON()
	printSummary()
}
