package sitemap

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	urlPackage "net/url"
	"sort"
	"sync"
	"time"
)

type Url string

type Flagger struct {
	mu    sync.Mutex
	flags map[Url]bool
}

func (s *Flagger) flag(url Url)  {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flags[url] = true
}

func (s *Flagger) isFlagged(url Url) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flags[url]
}

type Sitemap struct {
	mu sync.Mutex
	urls []Url
}

func (s *Sitemap) add(url Url)  {
	s.mu.Lock()
	defer s.mu.Unlock()
	var exists bool
	for _, u := range s.urls {
		if u == url {
			exists = true
		}
	}
	if !exists {
		s.urls = append(s.urls, url)
	}
}

func (s *Sitemap) sort()  {
	s.mu.Lock()
	defer s.mu.Unlock()
	sort.Slice(s.urls, func(i, j int) bool {
		return s.urls[i] < s.urls[j]
	})
}

func Create(url Url, depth int) ([]Url, error) {
	var wg sync.WaitGroup
	sitemap := Sitemap{sync.Mutex{}, []Url{}}
	results := make(chan Url)
	errors := make(chan error)
	extracted := Flagger{sync.Mutex{}, map[Url]bool{}}

	wg.Add(1)
	go crawl(&wg, url, depth, results, errors, &extracted)
	go fullfill(&sitemap, results, errors)

	wg.Wait()

	sitemap.sort()
	return sitemap.urls, nil
}

func fullfill(sitemap *Sitemap, results <-chan Url, errors <-chan error) {
	for {
		select {
		case url := <- results:
			sitemap.add(url)
			fmt.Printf("FOUND: %s\n", url)
		case err := <- errors:
			fmt.Printf("ERROR: %s\n", err)
		default:
		}
	}
}

func crawl(wg *sync.WaitGroup, url Url, depth int, results chan<- Url, errors chan<- error, extracted *Flagger) {
	defer wg.Done()

	parsedUrl, err := urlPackage.Parse(string(url))
	if err != nil {
		errors <- err
		return
	}

	urls, err := fetch(url)
	extracted.flag(url)
	if err != nil {
		errors <- err
		return
	}

	if depth > 1 {
		for _, u := range urls {
			if !extracted.isFlagged(u) {
				newUrl, err := urlPackage.Parse(string(u))
				if err != nil {
					errors <- err
					return
				}
				if newUrl.Scheme == "" {
					newUrl.Scheme = parsedUrl.Scheme
				}
				if newUrl.Host == "" {
					newUrl.Host = parsedUrl.Host
				}
				if newUrl.Host != parsedUrl.Host {
					continue
				}
				time.Sleep(100 * time.Millisecond)
				wg.Add(1)
				go crawl(wg, Url(newUrl.String()), depth - 1, results, errors, extracted)
			}
		}
	}

	results <- url
}

func fetch(url Url) ([]Url, error) {
	response, err := http.DefaultClient.Get(string(url))
	if err != nil {
		return nil, err
	}
	links, err := parseHtml(response.Body)
	if err != nil {
		return nil, err
	}
	var urls []Url
	urls = append(urls, links...)
	return urls, nil
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