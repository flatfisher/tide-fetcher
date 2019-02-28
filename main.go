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
	http.HandleFunc("/v1/tide", saveTideHandler)
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

func saveTideHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Project Id", http.StatusInternalServerError)
		return
	}
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		http.Error(w, "Cannot connect client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	if err := saveTide(ctx, client, tide); err != nil {
		http.Error(w, "Cannot save tide", http.StatusInternalServerError)
		return
	}

	//Create Json Response
	body, err := json.Marshal(tide)
	if err != nil {
		http.Error(w, "Cannot create Json Response", http.StatusInternalServerError)
		return
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
	queueID := "tide-request"
	key := os.Getenv("REQUEST_KEY")
	for _, v := range PORTS {
		go func(p Port) {
			url := makePath(p.Latitude, p.Longitude, key)
			_, err := createTask(projectID, locationID, queueID, url)
			if err != nil {
				log.Printf("Error: cannot create task %v", err)
			}
		}(v)
	}
	fmt.Fprint(w, "Create Tasks")
}

func makePath(lat float64, lon float64, key string) string {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	t := time.Now().In(jst)
	date := t.Format("20060102")
	return fmt.Sprintf("/v1/tide?date=%s&lat=%f&lon=%f&key=%s", date, lat, lon, key)
}
