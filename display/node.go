package display

import (
	"bytes"
	"html"
	"strings"
)

type page struct {
	Title string `json:"title"`
	Body  node   `json:"body"`
}

func page2html(buff *bytes.Buffer, p page) {
	buff.WriteString(`<!DOCTYPE html><html lang=fr>`)
	defer buff.WriteString(`</html>`)
	buff.WriteString(`<head>`)
	buff.WriteString(`<meta charset=utf-8>`)

	buff.WriteString(`<style>`)
	buff.Write(cssBytes)
	buff.WriteString(`</style>`)

	buff.WriteString(``)
	buff.WriteString(`<title>`)
	buff.WriteString(html.EscapeString(p.Title))
	buff.WriteString(`</title>`)
	buff.WriteString(`</head>`)

	p.Body.html(buff)
}

// Encode node in html.
func (n node) html(buff *bytes.Buffer) {
	buff.WriteByte('<')
	buff.WriteString(n.NodeName)
	if n.ID != "" {
		buff.WriteString(" id=")
		buff.WriteString(n.ID)
	}
	switch len(n.Classes) {
	case 0:
	case 1:
		buff.WriteString(" class=")
		buff.WriteString(n.Classes[0])
	default:
		buff.WriteString(` class="`)
		buff.WriteString(n.Classes[0])
		for _, c := range n.Classes[1:] {
			buff.WriteByte(' ')
			buff.WriteString(c)
		}
		buff.WriteByte('"')
	}
	for _, attr := range n.Attributes {
		buff.WriteByte(' ')
		buff.WriteString(attr)
	}
	buff.WriteByte('>')

	if n.Content != "" {
		buff.WriteString(html.EscapeString(n.Content))
	}

	for _, child := range n.Children {
		child.html(buff)
	}

	switch n.NodeName {
	case "img", "input":
	default:
		buff.WriteString("</")
		buff.WriteString(n.NodeName)
		buff.WriteByte('>')
	}
}

type node struct {
	// The type of the HTML element (div, span, a...)
	NodeName string `json:"nodename,omitempty"`
	// Identifier
	ID string `json:"id,omitempty"`
	// All classes, without starting dot.
	Classes []string `json:"classes,omitempty"`

	// All attributes
	Attributes []string `json:"attributes,omitempty"`

	Content  string `json:"content,omitempty"`
	Children []node `json:"children,omitempty"`
}

// New parent node.
func np(t string, children ...node) node {
	nodeName, id, classes := splitName(t)
	return node{
		NodeName: nodeName,
		ID:       id,
		Classes:  classes,
		Children: children,
	}
}

// New attributes and parent mode.
func nap(t string, attributes []string, children ...node) node {
	nodeName, id, classes := splitName(t)
	return node{
		NodeName:   nodeName,
		ID:         id,
		Classes:    classes,
		Attributes: attributes,
		Children:   children,
	}
}

// New node text
func nt(t, content string) node {
	nodeName, id, classes := splitName(t)
	return node{
		NodeName: nodeName,
		ID:       id,
		Classes:  classes,
		Content:  content,
	}
}

// Split the name in construction into some
func splitName(name string) (nodeName, id string, classes []string) {
	splited := strings.Split(name, ".")
	classes = splited[1:]
	nodeName, id, _ = strings.Cut(splited[0], "#")

	return
}
