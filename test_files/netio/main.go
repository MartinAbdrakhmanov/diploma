package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	sdk "github.com/MartinAbdrakhmanov/go-sdk"
)

func Handler(req sdk.Request) sdk.Response {
	city, ok := req.Body["city"].(string)
	if !ok || city == "" {
		city = "Moscow"
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	url := fmt.Sprintf("https://wttr.in/%s?format=j1", city)
	resp, err := client.Get(url)
	if err != nil {
		return sdk.Response{
			Status: 502,
			Body: map[string]interface{}{
				"error":   "Weather service unreachable",
				"details": err.Error(),
			},
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return sdk.Response{
			Status: 500,
			Body:   map[string]interface{}{"error": "Failed to read response body"},
		}
	}
	var weatherData map[string]interface{}
	if err := json.Unmarshal(body, &weatherData); err != nil {
		return sdk.Response{
			Status: 500,
			Body:   map[string]interface{}{"error": "Failed to parse weather data"},
		}
	}
	var tempC string
	if conditions, ok := weatherData["current_condition"].([]interface{}); ok && len(conditions) > 0 {
		if first, ok := conditions[0].(map[string]interface{}); ok {
			tempC = first["temp_C"].(string)
		}
	}

	return sdk.Response{
		Status: 200,
		Body: map[string]interface{}{
			"city":        city,
			"temperature": tempC + "°C",
			"info":        fmt.Sprintf("Current weather in %s is %s°C", city, tempC),
		},
	}
}

func main() {
	sdk.Run(Handler)
}
