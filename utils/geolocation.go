package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GeoLocationResponse struct {
	Status     string `json:"status"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	Query      string `json:"query"`
}

// Uses the ip-api to get geolocation information
func GetGeoLocationFromIP(ipAddress string) (string, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipAddress)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var geoLocationResponse GeoLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoLocationResponse); err != nil {
		return "", err
	}

	if geoLocationResponse.Status != "success" {
		return "", err
	}

	location := fmt.Sprintf("%s, %s, %s", geoLocationResponse.City, geoLocationResponse.RegionName, geoLocationResponse.Country)
	return location, nil
}
