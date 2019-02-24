package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/v1/tide", getTideHandler)
	http.HandleFunc("/v1/tide/tasks", taskHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Success")
}

func getTideHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/tide" {
		http.NotFound(w, r)
		return
	}

	key := r.URL.Query().Get("key")
	if key != os.Getenv("REQUEST_KEY") {
		http.Error(w, "Invalid API Key", http.StatusBadRequest)
		return
	}

	date := r.URL.Query().Get("date")
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")

	if (date == "") || (lat == "") || (lon == "") {
		http.Error(w, "Needs required propaties", http.StatusBadRequest)
		return
	}

	//Get Tide
	tide := getTideFromAPI(date, lat, lon)

	// Save Tide
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatalf("Set Firebase project ID via GOOGLE_CLOUD_PROJECT env variable.")
	}
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Cannot create client: %v", err)
	}
	defer client.Close()

	if err := saveTide(ctx, client, tide); err != nil {
		log.Fatalf("Cannot save tide: %v", err)
	}

	//Create Json Response
	body, err := json.Marshal(tide)
	if err != nil {
		log.Fatalf("Cannot create Json Response: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// Make requests for /v1/tide from Port List
func taskHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/tide/tasks" {
		http.NotFound(w, r)
		return
	}
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	locationID := "asia-northeast1"
	queueID := "tide"
	key := os.Getenv("REQUEST_KEY")
	for _, v := range PORTS {
		url := makePath(v.Latitude, v.Longitude, key)
		_, err := createTask(projectID, locationID, queueID, url)
		if err != nil {
			log.Fatalf("createTask: %v", err)
		}
	}
	fmt.Fprint(w, "Create Tasks")
}

func makePath(lat float64, lon float64, key string) string {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	t := time.Now().In(jst)
	date := t.Format("20060102")
	return fmt.Sprintf("/v1/tide?date=%s&lat%f&lon=%f&key=%s", date, lat, lon, key)
}
