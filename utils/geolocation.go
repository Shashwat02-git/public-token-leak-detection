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
	Message    string `json:"message"`
}

// GetGeoLocationFromIPs takes a list of IP addresses as strings and returns the first successful geolocation
func GetGeoLocationFromIPs(ipAddresses []string) (string, error) {
	var lastErr error

	for _, ipStr := range ipAddresses {
		url := fmt.Sprintf("http://ip-api.com/json/%s", ipStr)

		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch geolocation for %s: %v", ipStr, err)
			continue
		}

		var geoLocationResponse GeoLocationResponse
		if err := json.NewDecoder(resp.Body).Decode(&geoLocationResponse); err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("failed to decode response for %s: %v", ipStr, err)
			continue
		}
		resp.Body.Close()

		if geoLocationResponse.Status != "success" {
			lastErr = fmt.Errorf("API error for %s: %s", ipStr, geoLocationResponse.Message)
			continue
		}

		location := fmt.Sprintf("%s, %s, %s",
			geoLocationResponse.City,
			geoLocationResponse.RegionName,
			geoLocationResponse.Country)
		return location, nil
	}

	return "", fmt.Errorf("all IP lookups failed. Last error: %v", lastErr)
}
