package robotstxt

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

//go:embed ex-robots-monde-diplomatique.txt
var exRobotsMondeDiplomatique []byte

//go:embed ex-wikipedia-com.txt
var exWikipedia []byte

func TestParse(t *testing.T) {
	// Begin File from https://www.mediapart.fr/robots.txt (2022-09-05)
	assert.Equal(t, File{
		CrawlDelay: 1,
		SiteMap: []url.URL{
			url.URL{Scheme: "https", Host: "www.monde-diplomatique.fr", Path: "/sitemap.xml"},
		},
		Rules: []Rule{
			{true, "/squelettes/images/", []string{}, false},
			{false, "/squelettes/", []string{}, false},
			{false, "/squelettes-dist/", []string{}, false},
			{false, "/prive/", []string{}, false},
			{false, "/plugins/", []string{}, false},
			{true, "/local/cache-js/", []string{}, false},
			{true, "/local/cache-css/", []string{}, false},
			{true, "/local/", []string{}, false},
			{false, "/lib/", []string{}, false},
			{false, "/extensions/", []string{}, false},
			{false, "/ecrire/", []string{}, false},
		},
	}, Parse(exRobotsMondeDiplomatique))
}

func TestDevCutLines(t *testing.T) {
	assert.Equal(t, [][2]string{
		{"user-agent", "*"},
		{"disallow", "/cgi-bin/"},
		{"disallow", "/eur-lex/"},
		{"disallow", "/archives/"},
		{"disallow", "/youth/dissemination/"},
		{"disallow", "/youth/archive/"},
		{"disallow", "/youth/includes/"},
	}, parseLines([]byte(`# robots.txt for EUROPA httpd-80 production server
#
# last update on 20/06/2019
#
User-agent: *			# match any robot name
Disallow: /cgi-bin/		# don't allow robots into cgi-bin
Disallow: /eur-lex/		# don't index old Eurlex - 13/09/2006 Request from OPOCE (Mr O. Grossmann)
Disallow: /archives/	# don't index the archives

Disallow: /youth/dissemination/`+
		"\rDisallow: /youth/archive/"+
		"\r\nDisallow: /youth/includes/",
	)))
}

func TestFileAllow(t *testing.T) {
	file := Parse(exRobotsMondeDiplomatique)

	allow := func(urlString string, expected bool) {
		url, err := url.Parse(urlString)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, file.Allow(url), urlString)
	}

	allow("/", true)
	allow("/squelettes/", false)
	allow("/squelettes/yolo", false)
	allow("/squelettes/images/", true)
	allow("/squelettes/images/yolo", true)
}

func TestDefaultRobots(t *testing.T) {
	allow := func(urlString string) {
		url, err := url.Parse(urlString)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, DefaultRobots.Allow(url), urlString)
	}

	allow("/")
	allow("?fdbg")
	allow("/squelettes/")
	allow("/squelettes/yolo")
	allow("/squelettes/images/")
	allow("/squelettes/images/yolo")
}

func BenchmarkWikipedia(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse(exWikipedia)
	}
}
