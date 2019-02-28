package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
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

func getTideFromAPI(date string, lat string, lon string) (Tide, error) {
	var t Tide
	url := os.Getenv("API_URL")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return t, err
	}

	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("key", os.Getenv("API_KEY"))
	q.Add("date", date)
	q.Add("lat", lat)
	q.Add("lon", lon)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return t, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&t)
	if err != nil {
		return t, err
	}
	t.Location = &latlng.LatLng{Latitude: t.Latitude, Longitude: t.Longitude}
	d, err := getDate(t.DateString)
	if err != nil {
		return t, err
	}
	t.Date = d

	high, err := makeTides(t.HighTide, t.HighTideTime)
	if err != nil {
		return t, err
	}

	low, err := makeTides(t.LowTide, t.LowTideTime)
	if err != nil {
		return t, err
	}
	return t, nil
}

func saveTide(ctx context.Context, client *firestore.Client, t Tide) error {
	ref := client.Collection("tide").NewDoc()
	_, err := ref.Set(ctx, t)
	if err != nil {
		log.Printf("An error has occurred: %s", err)
	}
	return err
}

func makeTides(tide []string, time []string) ([]TideTime, error) {
	var tmp []TideTime
	for i, v := range tide {
		if v == "*" {
			break
		}
		t, err := getDate(time[i])
		if err != nil {
			return tmp, err
		}
		n, err := getInt(time[i])
		if err != nil {
			return tmp, err
		}
		tmp = append(tmp, TideTime{Level: n, Time: t})
	}
	return tmp, nil
}

func getDate(dateStr string) (time.Time, error) {
	RFC339 := "2006-01-02T15:04:05+09:00"
	time, err := time.Parse(RFC339, dateStr)
	if err != nil {
		return time, err
	}
	return time, nil
}

func getInt(intStr string) (int, error) {
	i, err := strconv.Atoi(intStr)
	if err != nil {
		return i, err
	}
	return i, nil
}
