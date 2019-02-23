package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"google.golang.org/genproto/googleapis/type/latlng"
)

// Tide is a response item
type Tide struct {
	DateString   string   `json:"date"`
	Port         string   `json:"pointName"`
	Code         string   `json:"pointCode"`
	Total        float64  `json:"totalTide"`
	Shiona       string   `json:"siona"`
	Average      float64  `json:"averageTide"`
	Moon         string   `json:"moonStatus"`
	Levels       []int    `json:"tideLevel"`
	Latitude     float64  `json:"lat"`
	Longitude    float64  `json:"lon"`
	HighTideTime []string `json:"highTideTime"`
	HighTide     []string `json:"highTide"`
	LowTideTime  []string `json:"lowTideTime"`
	LowTide      []string `json:"lowTide"`

	Location *latlng.LatLng
	High     []TideTime
	Low      []TideTime
	Date     time.Time
}

// TideTime is Time and tide level structure
type TideTime struct {
	Time  time.Time
	Level int
}

func getTide() Tide {
	url := os.Getenv("API_URL")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Cannot prepare request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("key", os.Getenv("API_KEY"))
	q.Add("lat", "44.35")
	q.Add("lon", "143.3666667")
	q.Add("date", "20190201")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Cannot request tide api: %v", err)
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var t Tide
	err = decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode tide body: %v", err)
	}
	t.Location = &latlng.LatLng{Latitude: t.Latitude, Longitude: t.Longitude}
	t.Date = getDate(t.DateString)
	t.High = makeTides(t.HighTide, t.HighTideTime)
	t.Low = makeTides(t.LowTide, t.LowTideTime)
	return t
}

func makeTides(tide []string, time []string) []TideTime {
	var tmp []TideTime
	for i, v := range tide {
		if v == "*" {
			break
		}
		tmp = append(tmp, TideTime{Level: getInt(v), Time: getDate(time[i])})
	}
	return tmp
}

func getDate(dateStr string) time.Time {
	RFC339 := "2006-01-02T15:04:05+09:00"
	time, err := time.Parse(RFC339, dateStr)
	if err != nil {
		log.Fatalf("Cannot parse string date: %v", err)
	}
	return time
}

func getInt(intStr string) int {
	i, err := strconv.Atoi(intStr)
	if err != nil {
		log.Fatalf("Cannot parse string date: %v", err)
	}
	return i
}
