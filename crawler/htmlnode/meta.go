package htmlnode

import (
	"golang.org/x/net/html/atom"
	"net/url"
	"strings"
)

type Meta struct {
	Langage     string
	Title       string
	Description string

	OpenGraph OpenGraph

	LinkedData [][]byte
}

// Visit the head to get some information stored into Meta.
func GetMeta(root Node) (meta Meta) {
	openGraph := make([][2]string, 0)

	root.Walk(func(node Node) bool {
		switch node.TagName {
		case atom.Html:
			meta.Langage = node.Attributes["lang"]
		case atom.Body:
			return true
		case atom.Title:
			if node.Text != "" {
				meta.Title = node.Text
			}
		case atom.Script:
			if node.Attributes["type"] == "application/ld+json" {
				if node.Text != "" {
					meta.LinkedData = append(meta.LinkedData, []byte(node.Text))
				}
			}
		case atom.Meta:
			content := node.Attributes["content"]
			if content == "" {
				return true
			}
			if node.Attributes["name"] == "description" {
				meta.Description = content
			} else if p := node.Attributes["property"]; strings.HasPrefix(p, "og:") {
				openGraph = append(openGraph, [2]string{p, content})
			}
		}
		return false
	})

	meta.OpenGraph = parseOpenGraph(openGraph)
	return
}

type OpenGraph struct {
	// Required properties:
	Title string
	Image OpenGraphMultimedia

	// Optional properties:
	Description string
	Local       string
	SiteName    string
}

// An image, a video or a audio.
type OpenGraphMultimedia struct {
	URL url.URL
}

func parseOpenGraph(meta [][2]string) (og OpenGraph) {
	for _, item := range meta {
		value := item[1]
		if value == "" {
			continue
		}
		switch item[0] {
		case "og:title":
			og.Title = value
		case "og:description":
			og.Description = value
		case "og:locale":
			og.Local = value
		case "og:site_name":
			og.SiteName = value

		case "og:image":
			if u, _ := url.Parse(value); u != nil {
				og.Image.URL = *u
			}
		}
	}
	return
}
