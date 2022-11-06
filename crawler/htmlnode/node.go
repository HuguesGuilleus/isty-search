package htmlnode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/url"
	"sort"
	"strings"
	"unicode"
)

const mimeLdJSON = "application/ld+json"

// The <html> element node.
type Root struct {
	// Root element attributes
	RootId         string
	RootClasses    []string
	RootAttributes map[string]string

	Meta Meta
	Head Node
	Body Node
}

// One html node, it can contain text or children. In case of pure text node,
// it do not contain Namespace, TagName and Attributes.
type Node struct {
	Namespace string
	TagName   atom.Atom

	Id         string
	Classes    []string
	Attributes map[string]string

	// All children. We use a slice insteed of a pointeur for parent,
	// first child... because of recursive data structure.
	Children []Node

	// Unescaped text.
	Text string
}

// From html content, get Root.
func Parse(data []byte) (*Root, error) {
	firstNode, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	htmlNode := searchNode(firstNode, atom.Html)
	headNode := searchNode(htmlNode, atom.Head)
	bodyNode := searchNode(htmlNode, atom.Body)

	rootId, rootClass, rootAttributes := convertAttributes(htmlNode.Attr)
	root := &Root{
		RootId:         rootId,
		RootClasses:    rootClass,
		RootAttributes: rootAttributes,
		Head:           convertOneNode(headNode),
		Body:           convertOneNode(bodyNode),
	}
	root.fillMeta()

	return root, nil
}

// Get the first html.Node white no Namespace and DataAtom == expected.
func searchNode(n *html.Node, expected atom.Atom) *html.Node {
	if n == nil {
		return nil
	}
	for {
		if n.DataAtom == expected && n.Namespace == "" {
			return n
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

func convertOneNode(srcNode *html.Node) Node {
	newNode := Node{
		Namespace: srcNode.Namespace,
		TagName:   srcNode.DataAtom,
	}
	newNode.Id, newNode.Classes, newNode.Attributes = convertAttributes(srcNode.Attr)

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

	if newNode.Namespace == "" &&
		newNode.TagName == atom.Script &&
		newNode.Attributes["type"] == mimeLdJSON {
		buff := recycler.Get()
		defer recycler.Recycle(buff)

		if err := json.Compact(buff, []byte(newNode.Text)); err == nil {
			newNode.Text = buff.String()
		}
	}

	return newNode
}

// Convert a html.Node and his children into a Node.
func convertNode(srcNode *html.Node, children *[]Node) {
	switch srcNode.Type {
	case html.TextNode:
		*children = append(*children, Node{Text: srcNode.Data})
	case html.ElementNode:
		if ignoreElementNode(srcNode) {
			return
		}
		*children = append(*children, convertOneNode(srcNode))
	}
}

// Return true if the element must be ignored.
// - Ignore style
// - Ignore script other than for linked data
func ignoreElementNode(srcNode *html.Node) bool {
	if srcNode.Namespace != "" {
		return false
	}

	if srcNode.DataAtom == atom.Style {
		return true
	} else if srcNode.DataAtom == atom.Script {
		for _, attr := range srcNode.Attr {
			if attr.Namespace == "" && attr.Key == "type" && attr.Val == mimeLdJSON {
				return false
			}
		}
		return true
	}

	return false
}

func convertAttributes(srcAttributes []html.Attribute) (id string, classes []string, newAttr map[string]string) {
	for i, attr := range srcAttributes {
		if attr.Namespace != "" {
			continue
		}
		switch key := strings.ToLower(attr.Key); key {
		case "class":
			classes = strings.FieldsFunc(attr.Val, unicode.IsSpace)
		case "id":
			id = attr.Val
		default:
			if newAttr == nil {
				newAttr = make(map[string]string, len(srcAttributes)-i)
			}
			newAttr[key] = attr.Val
		}
	}
	return
}

// From this document, get all url from anchor element.
// Filter url with protocol different to http or https.
func (root Root) GetURL(origin *url.URL) []*url.URL {
	foundedURL := make(map[string]bool, 0)
	root.Body.Visit(func(node Node) {
		if node.TagName == atom.A {
			if href := node.Attributes["href"]; href != "" {
				foundedURL[href] = true
			}
		}
	})

	sliceURL := make([]*url.URL, 0, len(foundedURL))
	for stringURL := range foundedURL {
		u, _ := origin.Parse(strings.Clone(stringURL))
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

// Call f on each node, and sub node.
// If f return true, do not walk into children.
func (node Node) Walk(f func(Node) bool) {
	toVisit := make([]Node, 1)
	toVisit[0] = node

	for len(toVisit) > 0 {
		n := toVisit[len(toVisit)-1]
		if f(n) {
			toVisit = toVisit[:len(toVisit)-1]
		} else {
			toVisit = append(toVisit[:len(toVisit)-1], n.Children...)
		}
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

		if id := node.Id; id != "" {
			buff.WriteString(" #")
			buff.WriteString(id)
		}

		for _, class := range node.Classes {
			buff.WriteString(" .")
			buff.WriteString(class)
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
