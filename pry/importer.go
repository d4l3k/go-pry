package pry

import (
	gcimporter "go/importer"
	"go/types"
)

func getImporter() types.Importer {
	return gcimporter.Default()
}
