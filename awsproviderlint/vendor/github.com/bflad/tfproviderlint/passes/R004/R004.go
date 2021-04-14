// Package R004 defines an Analyzer that checks for
// ResourceData.Set() calls using incompatible value types
package R004

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetcallexpr"
)

const Doc = `check for ResourceData.Set() calls using incompatible value types

The R004 analyzer reports incorrect types for a Set() call value.
The Set() function only supports a subset of basic types, slices and maps of that
subset of basic types, and the schema.Set type.`

const analyzerName = "R004"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		resourcedatasetcallexpr.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	sets := pass.ResultOf[resourcedatasetcallexpr.Analyzer].([]*ast.CallExpr)
	for _, set := range sets {
		if ignorer.ShouldIgnore(analyzerName, set) {
			continue
		}

		if len(set.Args) < 2 {
			continue
		}

		pos := set.Args[1].Pos()
		t := pass.TypesInfo.TypeOf(set.Args[1]).Underlying()

		if !isAllowedType(t) {
			pass.Reportf(pos, "%s: ResourceData.Set() incompatible value type: %s", analyzerName, t.String())
		}
	}

	return nil, nil
}

func isAllowedType(t types.Type) bool {
	switch t := t.(type) {
	default:
		return false
	case *types.Basic:
		return isAllowedBasicType(t)
	case *types.Interface:
		return true
	case *types.Map:
		switch k := t.Key().Underlying().(type) {
		default:
			return false
		case *types.Basic:
			if k.Kind() != types.String {
				return false
			}

			return isAllowedType(t.Elem().Underlying())
		}
	case *types.Named:
		return schema.IsNamedType(t, schema.TypeNameSet)
	case *types.Pointer:
		return isAllowedType(t.Elem())
	case *types.Slice:
		return isAllowedType(t.Elem().Underlying())
	}
}

var allowedBasicKindTypes = []types.BasicKind{
	types.Bool,
	types.Float32,
	types.Float64,
	types.Int,
	types.Int8,
	types.Int16,
	types.Int32,
	types.Int64,
	types.String,
	types.UntypedNil,
}

func isAllowedBasicType(b *types.Basic) bool {
	for _, allowedBasicKindType := range allowedBasicKindTypes {
		if b.Kind() == allowedBasicKindType {
			return true
		}
	}

	return false
}
