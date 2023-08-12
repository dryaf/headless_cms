package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dryaf/headless_cms/cache/memory_cache"
	"github.com/dryaf/headless_cms/client/storyblok"
	"github.com/joho/godotenv"
)

func loadEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := loadEnv("STORYBLOK_TOKEN", "")
	emptyCacheToken := loadEnv("STORYBLOK_EMPTY_CACHE_TOKEN", "")

	client := storyblok.NewClient(context.TODO(), token, emptyCacheToken, memory_cache.New(), &http.Client{})

	story, err := client.GetPageAsJSON(context.TODO(), "demo1", "published", "en")
	if err != nil {
		log.Fatalf("Failed to get story: %v", err)
	}
	fmt.Printf("Fetched Story from storyblok: %v\n\n", string(story))

	story, err = client.GetPageAsJSON(context.TODO(), "demo1", "published", "en")
	if err != nil {
		log.Fatalf("Failed to get story: %v", err)
	}
	fmt.Printf("Fetched Story from cache: %v", string(story))

}
