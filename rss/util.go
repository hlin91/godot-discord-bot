package rss

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

var presetCookies []*http.Cookie

func init() {
	presetCookies = append(presetCookies, &http.Cookie{
		Name:     "age_check",
		Value:    "1",
		Path:     "/",
		Domain:   `blog.fc2.com`,
		HttpOnly: true,
		Secure:   false,
		MaxAge:   0,
	})
}

type myJar struct {
	jar map[string][]*http.Cookie
}

func (p *myJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	log.Printf("The URL is : %s\n", u.String())
	log.Printf("The cookie being set is : %s\n", cookies)
	p.jar[u.Host] = cookies
}

func (p *myJar) Cookies(u *url.URL) []*http.Cookie {
	log.Printf("The URL is : %s\n", u.String())
	log.Printf("Cookie being returned is : %s\n", p.jar[u.Host])
	return append(p.jar[u.Host], presetCookies...)
}

func isImageFormat(s string) bool {
	formats := []string{".jpg", ".png", ".jpeg"}
	for _, fmt := range formats {
		if strings.HasSuffix(s, fmt) {
			return true
		}
	}
	return false
}

func isImageAttribute(s string) bool {
	// Element attributes that are known to contain image links
	var isImageAttr map[string]bool = map[string]bool{
		"href":     true,
		"src":      true,
		"url":      true,
		"srcset":   true,
		"data-src": true,
		"content":  true,
	}
	return isImageAttr[s]
}

func itemInList(list []*gofeed.Item, item *gofeed.Item) bool {
	for _, i := range list {
		if i.Title == item.Title {
			return true
		}
	}
	return false
}

// Recursively search for images in the html tree
func getImagesHelp(node *html.Node, dataType string, linksFound []string, n int) []string {
	if node == nil {
		return linksFound
	}
	if len(linksFound) >= n {
		return linksFound
	}
	if node.Type == html.ElementNode && node.Data == dataType {
		for _, a := range node.Attr {
			if isImageAttribute(a.Key) && isImageFormat(a.Val) && len(linksFound) < n {
				linksFound = append(linksFound, a.Val)
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		linksFound = getImagesHelp(c, dataType, linksFound, n)
	}
	return linksFound
}

// Get all nodes that comply with the given filter function
func getNodesByFunc(node *html.Node, filter func(*html.Node) bool, result []*html.Node) []*html.Node {
	if node == nil {
		return result
	}
	if filter(node) {
		result = append(result, node)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		result = getNodesByFunc(c, filter, result)
	}
	return result
}

// Run an extractor function to extract data from a list of nodes
func extractFromNodes(nodes []*html.Node, extractor func(*html.Node) string) []string {
	result := []string{}
	for _, n := range nodes {
		result = append(result, extractor(n))
	}
	return result
}

// Construct an http client with a working cookie jar
func httpClientWithCookieJar() *http.Client {
	// Set up the cookie jar so requests can be authenticated
	jar := &myJar{}
	jar.jar = map[string][]*http.Cookie{}
	client := &http.Client{}
	client.Jar = jar
	return client
}
