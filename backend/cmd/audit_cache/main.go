package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8080"
	endpoints := []struct {
		name string
		path string
	}{
		{"Initial Data", "/api/init"},
		{"Top Ranking", "/api/songs/ranking/top"},
		{"Songs List", "/api/songs?limit=20"},
	}

	fmt.Println("📊 --- CACHE VALIDATION AUDIT (5 Passes) ---")
	fmt.Printf("%-20s | %-10s | %-10s | %-10s | %-10s | %-10s\n", "Service", "Pass 1", "Pass 2", "Pass 3", "Pass 4", "Pass 5")
	fmt.Println("------------------------------------------------------------------------------------------")

	client := &http.Client{Timeout: 5 * time.Second}

	for _, ep := range endpoints {
		fmt.Printf("%-20s | ", ep.name)
		for i := 0; i < 5; i++ {
			start := time.Now()
			resp, err := client.Get(baseURL + ep.path)
			duration := time.Since(start)

			if err != nil {
				fmt.Printf("%-10s | ", "ERR")
				continue
			}
			resp.Body.Close()
			fmt.Printf("%-10s | ", duration.Round(time.Millisecond))
		}
		fmt.Println()
	}
	fmt.Println("------------------------------------------------------------------------------------------")
}
