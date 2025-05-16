package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (iPad; CPU OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:113.0) Gecko/20100101 Firefox/113.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7; rv:113.0) Gecko/20100101 Firefox/113.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:113.0) Gecko/20100101 Firefox/113.0",
	"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
}

func randomUserAgent() string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(userAgents)

	return userAgents[randNum]
}

func discoverLinks(response *http.Response, baseURL string) []string {
	if response != nil {
		defer response.Body.Close()
		doc, _ := goquery.NewDocumentFromReader(response.Body)
		foundUrls := []string{}

		if doc != nil {
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				res, _ := s.Attr("href")
				foundUrls = append(foundUrls, res)
			})
		}
		return foundUrls
	} else {
		return []string{}
	}
}

func getRequest(targetURL string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", randomUserAgent())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func checkRelative(href string, baseURL string) string {
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s%s", baseURL, href)
	} else {
		return href
	}
}

func resolveRelativeLinks(href string, baseURL string) (bool, string) {
	resultHref := checkRelative(href, baseURL)
	baseParse, _ := url.Parse(baseURL)
	resultParse, _ := url.Parse(resultHref)

	if baseParse != nil && resultParse != nil {
		if baseParse.Host == resultParse.Host {
			return true, resultHref
		} else {
			return false, ""
		}
	}
	return false, ""
}

var tokens = make(chan struct{}, 5)

func Crawl(targetURL string, baseURL string) []string {
	fmt.Println(targetURL)
	// semaphores, to limit the number of requests sent
	tokens <- struct{}{}

	resp, _ := getRequest(targetURL)
	<-tokens

	links := discoverLinks(resp, baseURL)
	foundUrls := []string{}

	for _, link := range links {
		ok, correctLink := resolveRelativeLinks(link, baseURL)
		if ok {
			if correctLink != "" {
				foundUrls = append(foundUrls, correctLink)
			}
		}
	}
	return foundUrls
}

func main() {
	worklist := make(chan []string)
	var n int
	n++
	baseDomain := "https://www.theguardian.com"

	go func() {
		worklist <- []string{"https://www.theguardian.com"}
	}()

	seen := make(map[string]bool)

	// to add new links to worklist and loop endlessly
	for ; n > 0; n-- {
		list := <-worklist
		for _, link := range list {
			if !seen[link] {
				seen[link] = true
				n++

				go func(link string, baseDomain string) {
					foundLinks := Crawl(link, baseDomain)

					if foundLinks != nil {
						worklist <- foundLinks
					}
				}(link, baseDomain)
			}
		}

	}
}
