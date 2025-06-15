package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Create request payload
	requestBody := map[string]interface{}{
		"description": "get orders with total order value",
		"system": "SystemA",
		"limit": 10,
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	
	// Send POST request to generate-query endpoint
	resp, err := http.Post(
		"http://localhost:8080/api/v1/generate-query", 
		"application/json", 
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	
	fmt.Println("Status:", resp.Status)
	fmt.Println("Response:")
	fmt.Println(string(body))
}