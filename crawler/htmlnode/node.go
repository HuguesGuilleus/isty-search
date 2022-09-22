package htmlnode

import (
	"bytes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
func Parse(data []byte) (*Node, []string, error) {
	root, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}

	return convertRoot(root), nil, nil
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
