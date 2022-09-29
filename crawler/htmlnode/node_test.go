package htmlnode

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
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
	receivedLine := root.PrintLines()

	expectedLine := (&Node{
		TagName:    atom.Html,
		Attributes: map[string]string{"lang": "en"},
		Children: []Node{
			Node{
				TagName: atom.Head,
				Children: []Node{
					Node{Text: "\n\t"},
					Node{
						TagName: atom.Title,
						Text:    "Hello World",
					},
					Node{Text: "\n\t"},
					Node{
						TagName: atom.Script,
						Text:    "\n\t\tconsole.log('Hello');\n\n\t",
					},
					Node{Text: "\n\t"},
					Node{
						TagName: atom.Style,
						Text:    "\n\t\t.yolo {\n\t\t\tcolor: red;\n\t\t}\n\n\t",
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
						Id:      "h1",
						Text:    "My First Heading",
					},
					Node{Text: "\n\t"},
					Node{
						TagName: atom.P,
						Classes: []string{"yolo", "swag"},
						Text:    "My first paragraph.",
					},
					Node{Text: "\n\t"},
					Node{
						TagName:    atom.Div,
						Attributes: map[string]string{"hidden": ""},
						Text:       "Hidden text!",
					},
					Node{Text: "\n\n\n\n"},
				},
			},
		},
	}).PrintLines()

	assert.EqualValues(t, expectedLine, receivedLine)
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
