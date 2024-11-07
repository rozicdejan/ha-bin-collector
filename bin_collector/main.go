package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	url            = "https://www.simbio.si/sl/moj-dan-odvoza-odpadkov"
	retryCount     = 3
	retryDelay     = 5 * time.Second
	requestTimeout = 10 * time.Second
	updateInterval = 15 * time.Minute
)

var (
	address string
	logger  *log.Logger
	// Use atomic operations or mutex consistently
	wasteData = struct {
		sync.RWMutex
		template TemplateData
		full     FullData
	}{}
)

// WasteSchedule represents the structure of a single JSON object in the response array
type WasteSchedule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Query   string `json:"query"`
	City    string `json:"city"`
	NextMKO string `json:"next_mko"`
	NextEmb string `json:"next_emb"`
	NextBio string `json:"next_bio"`
}

// TemplateData holds minimal data for HTML rendering
type TemplateData struct {
	MKOName string
	MKODate string
	EmbName string
	EmbDate string
	BioName string
	BioDate string
}

// FullData holds all data to be exposed via the API
type FullData struct {
	Name    string `json:"name"`
	Query   string `json:"query"`
	City    string `json:"city"`
	MKOName string `json:"mko_name"`
	MKODate string `json:"mko_date"`
	EmbName string `json:"emb_name"`
	EmbDate string `json:"emb_date"`
	BioName string `json:"bio_name"`
	BioDate string `json:"bio_date"`
}

func init() {
	// Initialize logger
	logger = log.New(os.Stdout, "[BIN-COLLECTOR] ", log.LstdFlags)

	// Get the address from the environment variable
	address = os.Getenv("ADDRESS")
	if address == "" {
		address = "začret 69"
		logger.Println("No ADDRESS environment variable set, using default:", address)
	}
}

func fetchDataWithRetry() error {
	var lastErr error
	for i := 0; i < retryCount; i++ {
		if err := fetchData(); err == nil {
			return nil
		} else {
			lastErr = err
			logger.Printf("Attempt %d failed: %v", i+1, err)
			time.Sleep(retryDelay)
		}
	}
	return fmt.Errorf("all retry attempts failed: %v", lastErr)
}

func fetchData() error {
	payload := []byte(fmt.Sprintf("action=simbioOdvozOdpadkov&query=%s", address))

	client := &http.Client{Timeout: requestTimeout}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var schedules []WasteSchedule
	if err := json.Unmarshal(body, &schedules); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if len(schedules) == 0 {
		return fmt.Errorf("no data received in the response")
	}

	firstSchedule := schedules[0]

	// Update the shared data with proper locking
	wasteData.Lock()
	wasteData.template = TemplateData{
		MKOName: "Mešani komunalni odpadki",
		MKODate: firstSchedule.NextMKO,
		EmbName: "Embalaža",
		EmbDate: firstSchedule.NextEmb,
		BioName: "Biološki odpadki",
		BioDate: firstSchedule.NextBio,
	}
	wasteData.full = FullData{
		Name:    firstSchedule.Name,
		Query:   firstSchedule.Query,
		City:    firstSchedule.City,
		MKOName: "Mešani komunalni odpadki",
		MKODate: firstSchedule.NextMKO,
		EmbName: "Embalaža",
		EmbDate: firstSchedule.NextEmb,
		BioName: "Biološki odpadki",
		BioDate: firstSchedule.NextBio,
	}
	wasteData.Unlock()

	return nil
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		logger.Printf("Error loading template: %v", err)
		return
	}

	wasteData.RLock()
	err = tmpl.Execute(w, wasteData.template)
	wasteData.RUnlock()

	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		logger.Printf("Error rendering template: %v", err)
	}
}

func apiDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wasteData.RLock()
	err := json.NewEncoder(w).Encode(wasteData.full)
	wasteData.RUnlock()

	if err != nil {
		http.Error(w, "Failed to encode data to JSON", http.StatusInternalServerError)
		logger.Printf("Error encoding JSON response: %v", err)
	}
}

func dataUpdater() {
	for {
		if err := fetchDataWithRetry(); err != nil {
			logger.Printf("Failed to update data: %v", err)
		}
		time.Sleep(updateInterval)
	}
}

func main() {
	logger.Println("Starting bin collector service...")

	// Initial data fetch
	if err := fetchDataWithRetry(); err != nil {
		logger.Printf("Initial data fetch failed: %v", err)
	}

	// Setup routes
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", dataHandler)
	http.HandleFunc("/api/data", apiDataHandler)

	// Start background updater
	go dataUpdater()

	// Start server
	addr := "0.0.0.0:8081"
	logger.Printf("Server running on http://%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}
