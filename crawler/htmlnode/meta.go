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

// Fill .Meta field.
func (root *Root) fillMeta() {
	root.Meta.Langage = root.RootAttributes["lang"]

	openGraph := make([][2]string, 0)
	root.Head.Visit(func(node Node) {
		switch node.TagName {
		case atom.Title:
			if node.Text != "" {
				root.Meta.Title = node.Text
			}
		case atom.Script:
			if node.Attributes["type"] == "application/ld+json" {
				if node.Text != "" {
					root.Meta.LinkedData = append(root.Meta.LinkedData, []byte(node.Text))
				}
			}
		case atom.Meta:
			content := node.Attributes["content"]
			if content == "" {
				return
			}
			if node.Attributes["name"] == "description" {
				root.Meta.Description = content
			} else if p := node.Attributes["property"]; strings.HasPrefix(p, "og:") {
				openGraph = append(openGraph, [2]string{p, content})
			}
		}
	})

	root.Meta.OpenGraph = parseOpenGraph(openGraph)
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
