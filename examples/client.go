package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Get fields endpoint
	resp, err := http.Get("http://localhost:8080/api/v1/fields")
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