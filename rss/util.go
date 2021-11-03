package rss

import (
	"strings"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

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
			if (a.Key == "href" || a.Key == "src") && isImageFormat(a.Val) && len(linksFound) < n {
				linksFound = append(linksFound, a.Val)
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		linksFound = getImagesHelp(c, dataType, linksFound, n)
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
