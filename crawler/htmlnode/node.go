package htmlnode

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/url"
	"sort"
)

// One html node, it can contain text or children. In case of pure text node,
// it do not contain Namespace, TagName and Attributes.
type Node struct {
	Namespace string
	TagName   atom.Atom

	Attributes map[string]string

	// All children. We use a slice insteed of a pointeur for parent,
	// first child... because of recursive data structure.
	Children []Node

	// Unescaped text.
	Text string
}

// From html content, get Node.
func Parse(data []byte) (*Node, error) {
	root, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return convertRoot(root), nil
}

// Search the root element and convert it in Node. So it remove doctype,
// some global comment, and the document node.
func convertRoot(n *html.Node) *Node {
	if n == nil {
		return nil
	}

	for {
		if n.Type == html.ElementNode && n.Namespace == "" && n.DataAtom == atom.Html {
			output := make([]Node, 0, 1)
			convertNode(n, &output)
			return &output[0]
		} else if n.FirstChild != nil {
			n = n.FirstChild
		} else if n.NextSibling != nil {
			n = n.NextSibling
		} else if n.Parent != nil {
			n = n.Parent.NextSibling
		} else {
			return nil
		}
	}
}

// Convert a html.Node and his children into a Node.
func convertNode(srcNode *html.Node, children *[]Node) {
	switch srcNode.Type {
	case html.TextNode:
		*children = append(*children, Node{Text: srcNode.Data})
	case html.ElementNode:
		newNode := Node{
			Namespace:  srcNode.Namespace,
			TagName:    srcNode.DataAtom,
			Attributes: convertAttributes(srcNode.Attr),
		}

		if first := srcNode.FirstChild; first != nil {
			if first == srcNode.LastChild && first.Type == html.TextNode {
				// only one text node
				newNode.Text = srcNode.FirstChild.Data
			} else {
				newNode.Children = make([]Node, 0)
				for child := srcNode.FirstChild; child != nil; child = child.NextSibling {
					convertNode(child, &newNode.Children)
				}
			}
		}

		*children = append(*children, newNode)
	}
}

func convertAttributes(srcAttributes []html.Attribute) map[string]string {
	if len(srcAttributes) == 0 {
		return nil
	}

	m := make(map[string]string, len(srcAttributes))

	for _, attr := range srcAttributes {
		if attr.Namespace != "" {
			m[attr.Namespace+":"+attr.Key] = attr.Val
		} else {
			m[attr.Key] = attr.Val
		}
	}

	return m
}

// From this document, get all url from anchor element.
// Filter url with protocol different of http or https.
func (node Node) GetURL(origin *url.URL) []*url.URL {
	foundedURL := make(map[string]bool, 0)
	node.Visit(func(node Node) {
		if node.TagName == atom.A {
			if href := node.Attributes["href"]; href != "" {
				foundedURL[href] = true
			}
		}
	})

	sliceURL := make([]*url.URL, 0, len(foundedURL))
	for stringURL := range foundedURL {
		u, _ := origin.Parse(stringURL)
		if u == nil {
			continue
		} else if u.Scheme != "https" && u.Scheme != "http" {
			continue
		}
		u.User = nil
		u.Fragment = ""
		u.ForceQuery = false
		u.RawQuery = u.Query().Encode()
		sliceURL = append(sliceURL, u)
	}

	return sliceURL
}

// Call f on each node, and sub node.
func (node Node) Visit(f func(Node)) {
	toVisit := make([]Node, 1)
	toVisit[0] = node

	for len(toVisit) > 0 {
		n := toVisit[len(toVisit)-1]
		f(n)
		toVisit = append(toVisit[:len(toVisit)-1], n.Children...)
	}
}

func (node *Node) PrintLines() []string {
	lines := make([]string, 0)
	node.printLines("", &lines)
	return lines
}

// Append recursively in lines each node description.
func (node *Node) printLines(tab string, lines *[]string) {
	buff := bytes.NewBufferString(tab)

	if node.TagName != 0 {
		if node.Namespace != "" {
			fmt.Fprintf(buff, "<%s:%s>", node.Namespace, node.TagName)
		} else {
			fmt.Fprintf(buff, "<%s>", node.TagName)
		}

		attributeNames := make([]string, 0, len(node.Attributes))
		for name := range node.Attributes {
			attributeNames = append(attributeNames, name)
		}
		sort.Strings(attributeNames)
		for _, name := range attributeNames {
			buff.WriteByte(' ')
			buff.WriteString(name)
			if value := node.Attributes[name]; value != "" {
				buff.WriteByte('=')
				buff.WriteString(value)
			}
		}
	}

	if len(node.Text) > 0 {
		fmt.Fprintf(buff, " '%s'", node.Text)
	}

	*lines = append(*lines, buff.String())

	for _, child := range node.Children {
		child.printLines(tab+"=", lines)
	}
}
