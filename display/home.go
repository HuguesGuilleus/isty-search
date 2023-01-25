package display

import (
	"bytes"
)

var home []byte = func() []byte {
	buff := bytes.Buffer{}
	page2html(&buff, page{
		Title: "ISTY Search",
		Body: np("body.home",
			np("div.home-top"),
			np("div.home-space1"),
			np("div.home-search",
				nt("div.home-search-title", "ISTY Search"),
				nap("form.home-search-form", []string{"action=/search"},
					nap(`input`, []string{
						"type=search",
						"name=q",
						`value=""`,
						`placeholder="Mots clés de recherche"`}),
					nap(`input`, []string{"type=submit", `value="⇢"`}),
				),
			),
			np("div.home-space2"),
			nt("footer.home-footer", "Hugues GUILLEUS, Projet ISTY-Search, 2022-2023"),
		),
	})
	return buff.Bytes()
}()
