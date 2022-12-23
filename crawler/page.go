package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"golang.org/x/net/html/atom"
	"net/url"
	"path"
	"strings"
	"time"
)

type Page struct {
	URL  url.URL
	Time time.Time

	// Content, on of the following filed.
	Error    string
	Html     *htmlnode.Root
	Robots   *robotstxt.File
	Redirect *url.URL
}

// Get all urls of the page
func (page *Page) GetURLs() map[crawldatabase.Key]*url.URL {
	if page.Html == nil {
		return nil
	}

	urls := make(map[crawldatabase.Key]*url.URL)
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
func getParentURL(urls map[crawldatabase.Key]*url.URL, src *url.URL) {
	// Source
	src.User = nil
	src.Fragment = ""
	src.ForceQuery = false
	src.RawQuery = src.Query().Encode()
	urls[crawldatabase.NewKeyURL(src)] = src

	// No query
	u := src.JoinPath("")
	u.RawQuery = ""
	urls[crawldatabase.NewKeyURL(u)] = u

	// Parent root
	u = src.JoinPath("")
	u.RawQuery = ""
	u.Path = "/"
	urls[crawldatabase.NewKeyURL(u)] = u
	for dir := path.Dir(src.Path); dir != "/"; dir = path.Dir(dir) {
		parent := u.JoinPath("/", dir+"/")
		urls[crawldatabase.NewKeyURL(parent)] = parent
	}

	// Port
	if newHost, _, cutted := strings.Cut(u.Host, ":"); cutted {
		u = u.JoinPath("")
		u.Host = newHost
		urls[crawldatabase.NewKeyURL(u)] = u
	}

	// Parent host
	host := strings.TrimSuffix(u.Host, ".")
	count := strings.Count(host, ".") - 1
	for i := 0; i < count; i++ {
		_, host, _ = strings.Cut(host, ".")
		u = u.JoinPath("")
		u.Host = host
		urls[crawldatabase.NewKeyURL(u)] = u
	}
	u = u.JoinPath("")
	u.Host = "www." + host
	urls[crawldatabase.NewKeyURL(u)] = u
}
