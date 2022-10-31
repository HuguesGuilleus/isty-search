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
	linkedData := `{"@context":"https:\/\/schema.org","@type":"Article","name":"Minijupe","url":"https:\/\/fr.wikipedia.org\/wiki\/Minijupe","sameAs":"http:\/\/www.wikidata.org\/entity\/Q230823","mainEntity":"http:\/\/www.wikidata.org\/entity\/Q230823","author":{"@type":"Organization","name":"Contributeurs aux projets Wikimedia"},"publisher":{"@type":"Organization","name":"Fondation Wikimedia, Inc.","logo":{"@type":"ImageObject","url":"https:\/\/www.wikimedia.org\/static\/images\/wmf-hor-googpub.png"}},"datePublished":"2005-11-21T14:16:45Z","dateModified":"2022-10-19T05:21:00Z","image":"https:/\/upload.wikimedia.org\/wikipedia\/commons\/b\/b0\/Minirock_%28Lack%29_Photo_Model_2.jpg","headline":"jupe tr\u00e8s courte, droite ou pliss\u00e9e, \u00ab dont la longueur ne devrait pas exc\u00e9der 10 cm sous les fesses pour justifier de cette appellation \u00bb"}`

	expected := &Root{
		RootId:         "root",
		RootClasses:    []string{"cr"},
		RootAttributes: map[string]string{"lang": "en"},

		Meta: Meta{
			Langage:    "en",
			Title:      "Hello World",
			LinkedData: [][]byte{[]byte(linkedData)},
		},

		Head: Node{
			TagName: atom.Head,
			Children: []Node{
				Node{Text: "\n\t"},
				Node{
					TagName: atom.Title,
					Text:    "Hello World",
				},
				Node{Text: "\n\t"},
				Node{Text: "\n\t"},
				Node{Text: "\n\t"},
				Node{
					TagName:    atom.Script,
					Attributes: map[string]string{"type": "application/ld+json"},
					Text:       linkedData,
				},
				Node{Text: "\n"},
			},
		},
		Body: Node{
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
	}

	received, err := Parse(exampleSimpleHtml)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, received)

	assert.Equal(t, expected.Body.PrintLines(), received.Body.PrintLines())
	assert.Equal(t, expected.Head.PrintLines(), received.Head.PrintLines())
	assert.Equal(t, expected, received)
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
