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
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

func randomUserAgent() string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

func discoverLinks(response *http.Response, baseURL string) []string {
	if response != nil {
		// was created to take the response that you get and create readable documents from it
		doc, _ := goquery.NewDocumentFromResponse(response)
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

func checkRelative(href string, baseUrl string) string {
	if strings.HasPrefix(href, "/") {
		return fmt.Sprintf("%s%s", baseUrl, href)
	} else {
		return href
	}

}

func resolveRelativeLinks(href string, baseUrl string) (bool, string) {
	resultHref := checkRelative(href, baseUrl)
	baseParse, _ := url.Parse(baseUrl)
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

var tokens = make(chan struct{}, 5) // creating a semaphore
// semaphores helps us to keep control of the number of simultaneous process that can be running at the same time

func Crawl(targetURL string, baseURL string) []string {
	fmt.Println(targetURL)
	tokens <- struct{}{}
	resp, _ := getRequest(targetURL)
	<-tokens
	links := discoverLinks(resp, baseURL)
	foundUrls := []string{}

	// range over the links
	for _, link := range links {
		ok, correctLink := resolveRelativeLinks(link, baseURL)
		if ok {
			if correctLink != "" {
				foundUrls = append(foundUrls, correctLink)
			}
		}
	}
	ParseHTML(resp)
	return foundUrls
}

func ParseHTML(response *http.Response) {
	// you can fill this
}

func main() {
	workList := make(chan []string) // Channels has links
	var n int                       // to make sure that the loop runs properly
	n++
	baseDomain := "https://www.theguardian.com"
	go func() { workList <- []string{"https://www.theguardian.com"} }() // the base domain is the first link to enter the list

	seen := make(map[string]bool) // to enable us be able to identify the links that we have seen and not have to repeat them

	for ; n > 0; n-- {
		list := <-workList

		for _, link := range list {
			if !seen[link] { // if link is not seen
				seen[link] = true
				n++
				go func(link string, baseURL string) {
					foundLinks := Crawl(link, baseDomain)
					if foundLinks != nil {
						workList <- foundLinks
					}
				}(link, baseDomain)
			}
		}
	}

}
