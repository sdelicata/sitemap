package main

import (
	"fmt"
	"github.com/sdelicata/sitemap/sitemap"
)

func main() {
	url := sitemap.Url("https://go.dev/")

	fmt.Println("Creating sitemap...")
	sm, err := sitemap.Create(url, 2)
	if err != nil {
		fmt.Printf("%s", err)
	}

	fmt.Println("==== SITEMAP ====")
	for _, u := range sm {
		fmt.Println(u)
	}
	fmt.Println("========")
	fmt.Printf("Count: %d", len(sm))
}
