package htmlnode

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

//go:embed example-maprimerenov.html
var exampleMaprimerenovHtml []byte

func TestGetMeta(t *testing.T) {
	root, err := Parse(exampleMaprimerenovHtml)
	assert.NoError(t, err)
	assert.Equal(t, Meta{
		Langage:     "fr",
		Title:       "MaPrimeRénov : avis du Défenseur des droits sur la dématérialisation | vie-publique.fr",
		Description: "La Défenseure des droits a été saisie de près de 500 réclamations rapportant les difficultés rencontrées par les usagers souhaitant bénéficier de MaPrimeRénov lors de leur démarche en ligne. Elle émet des recommandations à l'Agence nationale de l’habitat (Anah) en charge du dispositif d'aide à la rénovation des logements.",

		OpenGraph: OpenGraph{
			Title:       "MaPrimeRénov : la dématérialisation de la demande dénoncée par la Défenseure des droits",
			Description: "La Défenseure des droits a été saisie de près de 500 réclamations rapportant les difficultés rencontrées par les usagers souhaitant bénéficier de MaPrimeRénov lors de leur démarche en ligne. Elle émet des recommandations à l'Agence nationale de l’habitat (Anah) en charge du dispositif d'aide à la rénovation des logements.",
			SiteName:    "vie-publique.fr",
			Local:       "fr_FR",
			Image: OpenGraphMultimedia{
				URL: url.URL{
					Scheme: "https",
					Host:   "www.vie-publique.fr",
					Path:   "/sites/default/files/en_bref/image_principale/renovation-thermique.jpg",
				},
			},
		},

		LinkedData: [][]byte{[]byte(`{"@context":"https://schema.org","@graph":[{"@type":"Article","headline":"MaPrimeRénov : la dématérialisation de la demande dénoncée par la Défenseure des droits","about":["\u003Ca href=\u0022/relations-administration-usager\u0022 hreflang=\u0022fr\u0022\u003ERelations administration usager\u003C/a\u003E","\u003Ca href=\u0022/simplification-administrative\u0022 hreflang=\u0022fr\u0022\u003ESimplification administrative\u003C/a\u003E"],"description":"La Défenseure des droits a été saisie de près de 500 réclamations rapportant les difficultés rencontrées par les usagers souhaitant bénéficier de MaPrimeRénov lors de leur démarche en ligne. Elle émet des recommandations à l\u0026#039;Agence nationale de l’habitat (Anah) en charge du dispositif d\u0026#039;aide à la rénovation des logements.","image":{"@type":"ImageObject","representativeOfPage":"True","url":"https://www.vie-publique.fr/sites/default/files/styles/medium/public/en_bref/image_principale/renovation-thermique.jpg?itok=dL3LLnKW","width":"220","height":"138"},"datePublished":"2022-10-27T14:00:00+0200","dateModified":"2022-10-27T11:26:08+0200","isAccessibleForFree":"True","author":{"@type":"Organization","@id":"vie-publique.fr","name":"vie-publique.fr","url":"https://www.vie-publique.fr/","sameAs":["https://www.facebook.com/viepubliquefr/","http://twitter.com/LaDocFrancaise","https://www.youtube.com/channel/UCwYVByKhnWvujETeZYM87Ng?view_as=subscriber","https://www.instagram.com/ladocumentationfrancaise/"],"logo":{"@type":"ImageObject","width":"600","height":"140","url":"https://www.vie-publique.fr/sites/default/files/LOGO%20VP%20-%20Desktop_0.png"}},"publisher":{"@type":"Organization","@id":"vie-publique.fr","name":"vie-publique.fr","url":"https://www.vie-publique.fr/","sameAs":["https://www.facebook.com/viepubliquefr/","http://twitter.com/LaDocFrancaise","https://www.youtube.com/channel/UCwYVByKhnWvujETeZYM87Ng?view_as=subscriber","https://www.instagram.com/ladocumentationfrancaise/"],"logo":{"@type":"ImageObject","width":"600","height":"140","url":"https://www.vie-publique.fr/sites/default/files/LOGO%20VP%20-%20Desktop_0.png"}},"mainEntityOfPage":"https://www.vie-publique.fr/en-bref/286907-maprimerenov-avis-du-defenseur-des-droits-sur-la-dematerialisation"}]}`)},
	}, GetMeta(*root))
}
