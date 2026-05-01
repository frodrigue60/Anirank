package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Base URL for local development
	baseURL := "http://localhost:8080"
	
	endpoints := []struct {
		name string
		path string
	}{
		{"Initial Data (Init)", "/api/init"},
		{"Site Statistics", "/api/site-statistics"},
		{"Anime Catalog", "/api/animes?limit=20"},
		{"Songs Catalog", "/api/songs?limit=20"},
		{"Top Songs Ranking", "/api/songs/ranking/top"},
		{"Activity Feed", "/api/activity?limit=10"},
	}

	fmt.Println("\n🚀 --- ANIRANK PUBLIC API PERFORMANCE AUDIT ---")
	fmt.Printf("%-25s | %-15s | %-8s\n", "Service", "Latency", "Status")
	fmt.Println("------------------------------------------------------------")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, ep := range endpoints {
		url := baseURL + ep.path
		start := time.Now()
		resp, err := client.Get(url)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("%-25s | %-15s | %-8s (API Down?)\n", ep.name, "TIMEOUT", "ERR")
			continue
		}
		
		perfTag := "⚡ [FAST]"
		if duration > 50*time.Millisecond {
			perfTag = "🐢 [SLOW]"
		}
		if duration > 200*time.Millisecond {
			perfTag = "🛑 [CRITICAL]"
		}

		fmt.Printf("%-25s | %-15s | %-8d %s\n", ep.name, duration.Round(time.Microsecond), resp.StatusCode, perfTag)
		resp.Body.Close()
	}
	fmt.Println("------------------------------------------------------------\n")
}
