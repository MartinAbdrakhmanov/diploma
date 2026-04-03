package main

import (
	"net/http"
	"time"

	sdk "github.com/MartinAbdrakhmanov/go-sdk"
)

func Handler(req sdk.Request) sdk.Response {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("https://api.github.com/zen")
	if err != nil {
		return sdk.Response{
			Status: 502,
			Body: map[string]interface{}{
				"error":   "Failed to reach external API",
				"details": err.Error(),
			},
		}
	}
	defer resp.Body.Close()

	return sdk.Response{
		Status: 200,
		Body: map[string]interface{}{
			"message": "Outbound network call successful",
			"status":  resp.Status,
		},
	}
}

func main() {
	sdk.Run(Handler)
}
