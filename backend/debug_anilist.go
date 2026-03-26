package main

import (
	"context"
	"encoding/json"
	"fmt"
	"anirank/api/internal/infrastructure/anilist"
)

func main() {
	client := anilist.NewClient()
	ctx := context.Background()

	medias, err := client.GetMediaByIDs(ctx, []int{187901})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(medias) == 0 {
		fmt.Println("No media found")
		return
	}

	m := medias[0]
	fmt.Printf("Title: %s\n", m.Title.Romaji)
	fmt.Printf("Banner: %s\n", m.BannerImage)
	
	// Print as JSON for full debug
	b, _ := json.MarshalIndent(m, "", "  ")
	fmt.Println(string(b))
}
