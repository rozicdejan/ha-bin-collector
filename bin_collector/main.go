package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os" // Import for reading environment variables
	"sync"
	"time"
)

const (
	url            = "https://www.simbio.si/sl/moj-dan-odvoza-odpadkov"
	retryCount     = 3
	retryDelay     = 5 * time.Second
	requestTimeout = 10 * time.Second
)

var (
	address   string // The address input from Home Assistant environment
	wasteData = TemplateData{}
	fullData  = FullData{}
	mutex     = &sync.Mutex{}
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
	// Get the address from the environment variable, defaulting to "začret 69" if not set
	address = os.Getenv("ADDRESS")
	if address == "" {
		address = "začret 69"
	}
}

// fetchDataWithRetry makes an HTTP POST request to fetch the waste collection data with retry logic
func fetchDataWithRetry() {
	var err error
	for i := 0; i < retryCount; i++ {
		err = fetchData()
		if err == nil {
			return // Success
		}
		log.Printf("Attempt %d failed: %v", i+1, err)
		time.Sleep(retryDelay)
	}
	log.Println("All retry attempts failed. Serving with old data or fallback response.")
}

// fetchData makes an HTTP POST request to fetch the waste collection data
func fetchData() error {
	// Create the POST request payload
	payload := []byte(fmt.Sprintf("action=simbioOdvozOdpadkov&query=%s", address))

	// Create and configure the POST request
	client := &http.Client{Timeout: requestTimeout}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response as an array of WasteSchedule objects
	var schedules []WasteSchedule
	if err := json.Unmarshal(body, &schedules); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if len(schedules) == 0 {
		return fmt.Errorf("no data received in the response")
	}

	// Update the shared data with the first item in the array (adjust as needed)
	firstSchedule := schedules[0]
	mutex.Lock()
	wasteData = TemplateData{
		MKOName: "Mešani komunalni odpadki",
		MKODate: firstSchedule.NextMKO,
		EmbName: "Embalaža",
		EmbDate: firstSchedule.NextEmb,
		BioName: "Biološki odpadki",
		BioDate: firstSchedule.NextBio,
	}
	fullData = FullData{
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
	mutex.Unlock()
	return nil
}

// Serve dynamic HTML
func dataHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Error loading template: %v", err)
		return
	}
	if err := tmpl.Execute(w, wasteData); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
	}
}

// Serve JSON data via API
func apiDataHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fullData); err != nil {
		http.Error(w, "Failed to encode data to JSON", http.StatusInternalServerError)
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// Updates data every 15 minutes with retry logic
func dataUpdater() {
	for {
		fetchDataWithRetry()
		time.Sleep(15 * time.Minute)
	}
}

func main() {
	// Serve static files (e.g., for images or other assets if needed)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	go dataUpdater() // Runs the data updater in the background

	http.HandleFunc("/", dataHandler)
	http.HandleFunc("/api/data", apiDataHandler) // New API endpoint

	// Start the server
	fmt.Println("Server running on http://0.0.0.0:8081")
	log.Fatal(http.ListenAndServe("0.0.0.0:8081", nil))
}
