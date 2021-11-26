package main

import "fmt"

func main() {
	url := Url("https://rockenfolie.com/")

	fmt.Println("Creating sitemap...")
	sitemap, err := NewSitemap(url, 2)
	if err != nil {
		fmt.Printf("%s", err)
	}

	println("==== SITEMAP ====")
	for _, u := range sitemap {
		fmt.Println(u)
	}
	println("========")
	fmt.Printf("Count: %d", len(sitemap))
}
