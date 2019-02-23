package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Tide is a response item
type Tide struct {
	Port string `json:"pointName"`
}

func getTide() {
	url := os.Getenv("API_URL")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Cannot prepare request: %v", err)
	}

	q := req.URL.Query()
	q.Add("key", os.Getenv("API_KEY"))
	q.Add("lat", "44.35")
	q.Add("lon", "143.3666667")
	q.Add("date", "20190201")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Cannot request tide api: %v", err)
	}
	defer res.Body.Close()

	var t Tide
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode body: %v", err)
	}
}
