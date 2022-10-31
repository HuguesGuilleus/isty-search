package htmlnode

import (
	"golang.org/x/net/html/atom"
	"net/url"
	"strings"
	"unicode"
)

type Meta struct {
	Langage     string
	Title       string
	Description string

	NoFollow bool
	NoIndex  bool

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
			switch node.Attributes["name"] {
			case "description":
				root.Meta.Description = content
			case "robots":
				for _, value := range strings.FieldsFunc(content, isSpaceAndComa) {
					switch value {
					case "noindex":
						root.Meta.NoIndex = true
					case "nofollow":
						root.Meta.NoFollow = true
					case "none":
						root.Meta.NoIndex = true
						root.Meta.NoFollow = true
					}
				}
			default:
				if p := node.Attributes["property"]; strings.HasPrefix(p, "og:") {
					openGraph = append(openGraph, [2]string{p, content})
				}
			}

		}
	})

	root.Meta.OpenGraph = parseOpenGraph(openGraph)
}

func isSpaceAndComa(r rune) bool { return r == ',' || unicode.IsSpace(r) }

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
