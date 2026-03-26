package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"net/http"
	"os"
)

func main() {
	url := "https://s4.anilist.co/file/anilistcdn/media/anime/banner/187901-03I5T8QVX1FG.jpg"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Request creation failed: %v\n", err)
		return
	}
	req.Header.Set("User-Agent", "Anirank/1.0 (https://anirank.work)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP status failed: %s\n", resp.Status)
		return
	}

	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Content-Length: %d\n", resp.ContentLength)

	f, err := os.Create("test_banner.jpg")
	if err != nil {
		fmt.Printf("File creation failed: %v\n", err)
		return
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		fmt.Printf("Copy failed: %v\n", err)
		return
	}

	fmt.Printf("Downloaded %d bytes\n", n)

	// Re-read file to decode
	f.Seek(0, 0)
	_, _, err = image.Decode(f)
	if err != nil {
		fmt.Printf("Decode failed: %v\n", err)
	} else {
		fmt.Println("Successfully decoded image")
	}
}
