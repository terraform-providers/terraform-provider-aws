package exhaustive

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

func checkMapLiterals(pass *analysis.Pass, inspect *inspector.Inspector, comments map[*ast.File]ast.CommentMap) {
	for _, f := range pass.Files {
		for _, d := range f.Decls {
			gen, ok := d.(*ast.GenDecl)
			if !ok {
				continue
			}
			if gen.Tok != token.VAR {
				continue // map literals have to be declared as "var"
			}
			for _, s := range gen.Specs {
				valueSpec := s.(*ast.ValueSpec)
				for idx, name := range valueSpec.Names {
					obj := pass.TypesInfo.Defs[name]
					if obj == nil {
						continue
					}

					mapType, ok := obj.Type().(*types.Map)
					if !ok {
						continue
					}

					keyType, ok := mapType.Key().(*types.Named)
					if !ok {
						continue
					}
					keyPkg := keyType.Obj().Pkg()
					if keyPkg == nil {
						// Doc comment: nil for labels and objects in the Universe scope.
						// This happens for the `error` type, for example.
						// Continuing would mean that ImportPackageFact panics.
						continue
					}

					var enums enumsFact
					if !pass.ImportPackageFact(keyPkg, &enums) {
						// Can't do anything further.
						continue
					}

					enumMembers, ok := enums.Entries[keyType.Obj().Name()]
					if !ok {
						// Key type is not a known enum.
						continue
					}

					// Check comments for the ignore directive.

					var allComments ast.CommentMap
					if cm, ok := comments[f]; ok {
						allComments = cm
					} else {
						allComments = ast.NewCommentMap(pass.Fset, f, f.Comments)
						comments[f] = allComments
					}

					genDeclComments := allComments.Filter(gen)
					genDeclIgnore := false
					for _, group := range genDeclComments.Comments() {
						if containsIgnoreDirective(group.List) && gen.Lparen == token.NoPos && len(gen.Specs) == 1 {
							genDeclIgnore = true
							break
						}
					}
					if genDeclIgnore {
						continue
					}

					if (valueSpec.Doc != nil && containsIgnoreDirective(valueSpec.Doc.List)) ||
						(valueSpec.Comment != nil && containsIgnoreDirective(valueSpec.Comment.List)) {
						continue
					}

					samePkg := keyPkg == pass.Pkg
					checkUnexported := samePkg

					hitlist := hitlistFromEnumMembers(enumMembers, checkUnexported)
					if len(hitlist) == 0 {
						// can happen if external package and enum consists only of
						// unexported members
						continue
					}

					if !(len(valueSpec.Values) > idx) {
						continue // no value for name
					}
					comp, ok := valueSpec.Values[idx].(*ast.CompositeLit)
					if !ok {
						continue
					}
					for _, el := range comp.Elts {
						kvExpr, ok := el.(*ast.KeyValueExpr)
						if !ok {
							continue
						}
						e := astutil.Unparen(kvExpr.Key)
						if samePkg {
							ident, ok := e.(*ast.Ident)
							if !ok {
								continue
							}
							delete(hitlist, ident.Name)
						} else {
							selExpr, ok := e.(*ast.SelectorExpr)
							if !ok {
								continue
							}
							delete(hitlist, selExpr.Sel.Name)
						}
					}

					if len(hitlist) > 0 {
						reportMapLiteral(pass, name, samePkg, keyType, hitlist)
					}
				}
			}
		}
	}
}

func reportMapLiteral(pass *analysis.Pass, mapVarIdent *ast.Ident, samePkg bool, enumType *types.Named, missingMembers map[string]struct{}) {
	missing := make([]string, 0, len(missingMembers))
	for m := range missingMembers {
		missing = append(missing, m)
	}
	sort.Strings(missing)

	pass.Report(analysis.Diagnostic{
		Pos:     mapVarIdent.Pos(),
		Message: fmt.Sprintf("missing keys in map %s of key type %s: %s", mapVarIdent.Name, enumTypeName(enumType, samePkg), strings.Join(missing, ", ")),
	})
}
