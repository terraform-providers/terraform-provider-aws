package ruleguard

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/quasilyte/go-ruleguard/dslgen"
)

type dslImporter struct {
	fallback types.Importer
}

func newDSLImporter() *dslImporter {
	return &dslImporter{fallback: importer.Default()}
}

func (i *dslImporter) Import(path string) (*types.Package, error) {
	switch path {
	case "github.com/quasilyte/go-ruleguard/dsl/fluent":
		return i.importDSL(path, dslgen.Fluent)

	default:
		return i.fallback.Import(path)
	}
}

func (i *dslImporter) importDSL(path string, src []byte) (*types.Package, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "dsl.go", src, 0)
	if err != nil {
		return nil, err
	}
	var typecheker types.Config
	var info types.Info
	return typecheker.Check(path, fset, []*ast.File{f}, &info)
}
