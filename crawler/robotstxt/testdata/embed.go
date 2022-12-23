package robotstxttestdata

import (
	_ "embed"
)

//go:embed ex-robots-monde-diplomatique.txt
var MondeDiplomatique []byte

//go:embed ex-wikipedia-com.txt
var Wikipedia []byte
