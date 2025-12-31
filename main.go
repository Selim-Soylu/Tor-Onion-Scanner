package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v3"
)

const (
	torProxy    = "127.0.0.1:9050"
	timeout     = 90 * time.Second
	outputDir   = "output"
	screenDir   = "screenshots"
	reportFile  = "scan_report.log"
	workerCount = 5
)

type Config struct {
	Targets []string `yaml:"targets"`
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Kullanım: go run main.go targets.yaml")
		return
	}

	os.MkdirAll(outputDir, 0755)
	os.MkdirAll(screenDir, 0755)

	logFile, _ := os.OpenFile(reportFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	log.SetOutput(logFile)

	log.Println("[INFO] Checking Tor connection...")

	httpClient, err := torHTTPClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := checkTor(httpClient); err != nil {
		log.Fatal("[FATAL] Tor connection failed:", err)
	}

	log.Println("[INFO] Tor connection verified")

	cfg, err := loadYAML(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	urlChan := make(chan string)
	var wg sync.WaitGroup

	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go worker(i, urlChan, httpClient, &wg)
	}

	for _, url := range cfg.Targets {
		url = strings.TrimSpace(url)
		if url != "" {
			urlChan <- url
		}
	}

	close(urlChan)
	wg.Wait()

	fmt.Println("Tarama tamamlandı.")
}

func loadYAML(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func worker(id int, urls <-chan string, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urls {
		fmt.Println("[WORKER", id, "]", url)

		if err := fetchHTML(client, url); err != nil {
			log.Println("[ERR]", url, "HTML ->", err)
			continue
		}

		// 🔴 HER URL İÇİN AYRI CHROME CONTEXT (KRİTİK FIX)
		chromeCtx, cancel := torChromeContext()

		if err := takeScreenshot(chromeCtx, url); err != nil {
			log.Println("[ERR]", url, "SCREENSHOT ->", err)
			cancel()
			continue
		}

		cancel()
		log.Println("[INFO]", url, "-> SUCCESS")
	}
}

func torHTTPClient() (*http.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", torProxy, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
		Timeout:   timeout,
	}, nil
}

func fetchHTML(client *http.Client, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(outputDir+"/"+sanitize(url)+".html", body, 0644)
}

func torChromeContext() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath("/usr/bin/chromium"),
		chromedp.ProxyServer("socks5://"+torProxy),
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	return ctx, func() {
		ctxCancel()
		allocCancel()
	}
}

func takeScreenshot(ctx context.Context, url string) error {
	var buf []byte

	c, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := chromedp.Run(c,
		chromedp.Navigate(url),
		chromedp.Sleep(8*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return err
	}

	return os.WriteFile(screenDir+"/"+sanitize(url)+".png", buf, 0644)
}

func sanitize(url string) string {
	url = strings.ReplaceAll(url, "http://", "")
	url = strings.ReplaceAll(url, "https://", "")
	url = strings.ReplaceAll(url, "/", "_")
	return url
}

func checkTor(client *http.Client) error {
	resp, err := client.Get("https://check.torproject.org/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Congratulations") {
		return fmt.Errorf("Tor IP doğrulanamadı")
	}
	return nil
}
