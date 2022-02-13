package rss

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

type myJar struct {
	jar map[string][]*http.Cookie
}

func (p *myJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	fmt.Printf("The URL is : %s\n", u.String())
	fmt.Printf("The cookie being set is : %s\n", cookies)
	p.jar[u.Host] = cookies
}

func (p *myJar) Cookies(u *url.URL) []*http.Cookie {
	fmt.Printf("The URL is : %s\n", u.String())
	fmt.Printf("Cookie being returned is : %s\n", p.jar[u.Host])
	return p.jar[u.Host]
}

// Element attributes that are known to contain image links
var isImageAttr map[string]bool = map[string]bool{
	"href":     true,
	"src":      true,
	"url":      true,
	"srcset":   true,
	"data-src": true,
}

func getImageFormats() []string {
	return []string{".jpg", ".png", ".jpeg"}
}

func isImageFormat(s string) bool {
	formats := getImageFormats()
	for _, fmt := range formats {
		if strings.HasSuffix(s, fmt) {
			return true
		}
	}
	return false
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
			if isImageAttr[a.Key] && isImageFormat(a.Val) && len(linksFound) < n {
				linksFound = append(linksFound, a.Val)
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		linksFound = getImagesHelp(c, dataType, linksFound, n)
	}
	// Some links found will begin with "//" without specifying the protocol
	// This causes issues so we will trim it from the string
	for i, s := range linksFound {
		linksFound[i] = strings.TrimLeft(s, "/")
	}
	return linksFound
}

// Get all the nodes that match a given class name
func getNodesByClass(node *html.Node, class string, result []*html.Node) []*html.Node {
	if node == nil {
		return result
	}
	for _, a := range node.Attr {
		if a.Key == "class" {
			classes := strings.Split(a.Val, ",")
			for _, c := range classes {
				if c == class {
					result = append(result, node)
					break
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		result = getNodesByClass(c, class, result)
	}
	return result
}

func httpClientWithCookieJar() *http.Client {
	// Set up the cookie jar so requests can be authenticated
	jar := &myJar{}
	jar.jar = map[string][]*http.Cookie{}
	client := &http.Client{}
	client.Jar = jar
	return client
}
