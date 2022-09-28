package htmlnode

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/url"
)

// One node, it can contain text or is a container node with a tag,
// and maybe attributes and children.
type Node struct {
	// Unescaped text.
	Text string

	Namespace  string
	TagName    atom.Atom
	Attributes []html.Attribute
	// All children. We use a slice insteed of a pointeur for parent,
	// first child... because of recursive data structure.
	Children []Node
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
			Attributes: srcNode.Attr,
		}

		if srcNode.FirstChild != nil {
			newNode.Children = make([]Node, 0)
			for child := srcNode.FirstChild; child != nil; child = child.NextSibling {
				convertNode(child, &newNode.Children)
			}
		}

		*children = append(*children, newNode)
	}
}

// From this document, get all url from anchor element.
// Filter url with protocol different of http or https.
func (node Node) GetURL(origin *url.URL) []*url.URL {
	foundedURL := make(map[string]bool, 0)
	node.Visit(func(node Node) {
		if node.TagName == atom.A {
			for _, attr := range node.Attributes {
				if attr.Namespace == "" && attr.Key == "href" {
					foundedURL[attr.Val] = true
				}
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
	s := tab
	if len(node.Text) > 0 {
		s += fmt.Sprintf("'%s'", node.Text)
	} else {
		if node.Namespace != "" {
			s += fmt.Sprintf("<%s:%s>", node.Namespace, node.TagName)
		} else {
			s += fmt.Sprintf("<%s>", node.TagName)
		}
		for _, attr := range node.Attributes {
			s += " "
			if attr.Namespace != "" {
				s += attr.Namespace + ":"
			}
			s += attr.Key
			if attr.Val != "" {
				s += "=" + attr.Val
			}
		}
	}
	*lines = append(*lines, s)

	for _, child := range node.Children {
		child.printLines(tab+"=", lines)
	}
}
