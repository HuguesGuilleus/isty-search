package htmlnode

import (
	_ "embed"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/url"
	"sort"
	"testing"
)

var (
	//go:embed example-simple.html
	exampleSimpleHtml []byte
	//go:embed example-url.html
	exampleURLHtml []byte
)

func TestParse(t *testing.T) {
	root, err := Parse(exampleSimpleHtml)
	assert.NoError(t, err)
	receivedLine := make([]string, 0)
	root.print("", &receivedLine)

	expectedLine := make([]string, 0)
	(&Node{
		TagName:    atom.Html,
		Attributes: []html.Attribute{{"", "lang", "en"}},
		Children: []Node{
			Node{
				TagName: atom.Head,
				Children: []Node{
					Node{Text: "\n\t"},
					Node{
						TagName: atom.Title,
						Children: []Node{
							Node{Text: "Hello World"},
						},
					},
					Node{Text: "\n\t"},
					Node{
						TagName: atom.Script,
						Children: []Node{
							Node{Text: "\n\t\tconsole.log('Hello');\n\n\t"},
						},
					},
					Node{Text: "\n\t"},
					Node{
						TagName: atom.Style,
						Children: []Node{
							Node{Text: "\n\t\t.yolo {\n\t\t\tcolor: red;\n\t\t}\n\n\t"},
						},
					},
					Node{Text: "\n"},
				},
			},
			Node{Text: "\n\n"},
			Node{
				TagName: atom.Body,
				Children: []Node{
					Node{Text: "\n\t"},
					Node{
						TagName: atom.H1,
						Children: []Node{
							Node{Text: "My First Heading"},
						},
					},
					Node{Text: "\n\t"},
					Node{
						TagName: atom.P,
						Attributes: []html.Attribute{
							{"", "class", "yolo"},
						},
						Children: []Node{
							Node{Text: "My first paragraph."},
						},
					},
					Node{Text: "\n\n\n\n"},
				},
			},
		},
	}).print("", &expectedLine)

	assert.EqualValues(t, expectedLine, receivedLine)
}

// Append recursively in lines each node description.
func (node *Node) print(tab string, lines *[]string) {
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
		child.print(tab+"=", lines)
	}
}

func TestGetURL(t *testing.T) {
	mustParse := func(s string) *url.URL {
		u, err := url.Parse(s)
		if err != nil {
			panic(err)
		}
		return u
	}

	root, err := Parse(exampleURLHtml)
	assert.NoError(t, err)

	receivedURL := root.GetURL(mustParse("https://example.com/dir/subdir/file.html"))
	sort.Slice(receivedURL, func(i, j int) bool {
		return receivedURL[i].String() < receivedURL[j].String()
	})

	assert.Equal(t, []*url.URL{
		mustParse("https://example.com/"),
		mustParse("https://example.com/dir/"),
		mustParse("https://example.com/dir/subdir/file.html"),
		mustParse("https://example.com/dir/subdir/file.html?a=1&b=2"),
		mustParse("https://example.com/swag"),
		mustParse("https://github.com/"),
	}, receivedURL)
}
