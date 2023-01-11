package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"github.com/HuguesGuilleus/isty-search/keys"
	"golang.org/x/net/html/atom"
	"net/url"
	"strings"
)

type Page struct {
	URL url.URL

	// Content, on of the following filed.
	Html   *htmlnode.Root
	Robots *robotstxt.File
}

// Get all urls of the page
func (page *Page) GetURLs() map[keys.Key]*url.URL {
	if page.Html == nil {
		return nil
	}

	urls := make(map[keys.Key]*url.URL)
	page.Html.Body.Visit(func(node htmlnode.Node) {
		if node.TagName == atom.A {
			if href := node.Attributes["href"]; href != "" {
				u, _ := page.URL.Parse(href)
				if u == nil || (u.Scheme != "https" && u.Scheme != "http") {
					return
				}
				getParentURL(urls, u)
			}
		}
	})

	return urls
}

// Get all parent of the srouce url (no query, path parent, and path host.)
func getParentURL(urls map[keys.Key]*url.URL, src *url.URL) {
	// Source
	cleanURL(src)
	src.Host = strings.TrimSuffix(src.Host, ".")
	if key := keys.NewURL(src); urls[key] != nil {
		return
	} else {
		urls[key] = src
	}

	// No query
	if src.RawQuery != "" {
		u := cloneURL(src)
		u.RawQuery = ""
		urls[keys.NewURL(u)] = u
	}

	// Parent root
	u := cloneURL(src)
	u.RawQuery = ""
	for i := len(u.Path) - 1; i >= 0; i-- {
		if u.Path[i] == '/' {
			u = cloneURL(u)
			u.Path = u.Path[:i+1]
			urls[keys.NewURL(u)] = u
		}
	}
	u = cloneURL(u)
	u.Path = "/"

	// Port
	if newHost, _, cutted := strings.Cut(u.Host, ":"); cutted {
		u.Host = newHost
		urls[keys.NewURL(u)] = u
		u = cloneURL(u)
	}

	// Parent host
	count := strings.Count(u.Host, ".") - 1
	for i := 0; i < count; i++ {
		_, u.Host, _ = strings.Cut(u.Host, ".")
		urls[keys.NewURL(u)] = u
		u = cloneURL(u)
	}
	u.Host = "www." + u.Host
	urls[keys.NewURL(u)] = u
}

func cloneURL(src *url.URL) *url.URL {
	u := *src
	return &u
}

// Clean url (remove user+fragment, sort query)
func cleanURL(u *url.URL) {
	u.User = nil
	u.Fragment = ""
	u.ForceQuery = false
	u.RawQuery = u.Query().Encode()
}
