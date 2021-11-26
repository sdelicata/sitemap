package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"sort"
)

type Url string

var sitemap []Url
var extracted map[Url]bool
var websiteHost string
var limit int
var cnt int

func NewSitemap(websiteUrl Url, l int) ([]Url, error) {
	extracted = map[Url]bool{}

	parsedWebsiteUrl, err := url.Parse(string(websiteUrl))
	if err != nil {
		return nil, err
	}
	websiteHost = parsedWebsiteUrl.Host

	limit = l

	cnt = 0

	record(websiteUrl)
	sitemap = unique(sitemap)
	sort.Slice(sitemap, func(i, j int) bool {
		return sitemap[i] < sitemap[j]
	})

	return sitemap, nil
}

func record(pageUrl Url) {
	fmt.Printf("Recording %s\n", pageUrl)
	links, err := extractLinks(pageUrl)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	var results []Url
	for _, l := range links {
		u, err := url.Parse(string(l))
		if err != nil {
			fmt.Printf("%s", err)
		}
		if u.Host != websiteHost {
			continue
		}
		newUrl := Url(u.String())
		results = append(results, newUrl)
	}
	sitemap = append(sitemap, unique(results)...)
	extracted[pageUrl] = true

	cnt++
	if cnt >= limit {
		return
	}
	for _, r := range results {
		if extracted[r] != true && cnt < limit {
			record(r)
		}
	}
}

func extractLinks(pageUrl Url) ([]Url, error) {
	response, err := http.DefaultClient.Get(string(pageUrl))
	if err != nil {
		return nil, err
	}
	links, err := parseHtml(response.Body)
	if err != nil {
		fmt.Printf("%s", err)
		return nil, err
	}
	return unique(links), nil
}

func parseHtml(r io.Reader) ([]Url, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var urls []Url
	var rec func(*html.Node)
	rec = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					urls = append(urls, Url(attr.Val))
				}
			}
		}
		if n.FirstChild != nil {
			rec(n.FirstChild)
		}
		if n.NextSibling != nil {
			rec(n.NextSibling)
		}
	}
	rec(root)

	return urls, nil
}

func unique(urls []Url) []Url {
	exists := map[Url]bool{}
	var results []Url
	for _, u := range urls {
		if exists[u] != true {
			exists[u] = true
			results = append(results, u)
		}
	}
	return results
}