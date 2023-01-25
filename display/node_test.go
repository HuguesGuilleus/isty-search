package display

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNodeHtml(t *testing.T) {
	defer func(oldCssByte []byte) { cssBytes = oldCssByte }(cssBytes)
	cssBytes = []byte("/*CSS*/")

	buff := bytes.Buffer{}
	page2html(&buff, page{
		Title: "<> Title!",
		Body: np("body#id.aaa.bbb",
			nt("div.home-search-title", "ISTY Search"),
			np("div",
				nap("input", []string{
					`type=search`,
					`value=""`,
					`placeholder="Mots clés de recherche"`,
				}),
			),
		),
	})

	assert.Equal(t, []string{
		`<!DOCTYPE html>`,
		`<html lang=fr>`,

		`<head>`,
		`<meta charset=utf-8>`,
		`<style>`, `/*CSS*/</style>`,
		`<title>`, `&lt;&gt; Title!</title>`,
		`</head>`,

		`<body id=id class="aaa bbb">`,
		`<div class=home-search-title>`,
		`ISTY Search</div>`,
		`<div>`,
		`<input type=search value="" placeholder="Mots clés de recherche">`,
		`</div>`,

		`</body>`,

		`</html>`, ``,
	}, strings.SplitAfter(buff.String(), ">"))
}

func TestSplitName(t *testing.T) {
	nodeName, id, classes := splitName("div#id.c1.c2")
	assert.Equal(t, "div", nodeName)
	assert.Equal(t, "id", id)
	assert.Equal(t, []string{"c1", "c2"}, classes)

	nodeName, id, classes = splitName("div#id")
	assert.Equal(t, "div", nodeName)
	assert.Equal(t, "id", id)
	assert.Empty(t, classes)

	nodeName, id, classes = splitName("div")
	assert.Equal(t, "div", nodeName)
	assert.Equal(t, "", id)
	assert.Empty(t, classes)

	nodeName, id, classes = splitName("div.c1.c2")
	assert.Equal(t, "div", nodeName)
	assert.Equal(t, "", id)
	assert.Equal(t, []string{"c1", "c2"}, classes)
}
