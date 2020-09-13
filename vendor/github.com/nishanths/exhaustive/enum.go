package exhaustive

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type enums map[string][]string // enum type name -> enum member names

func findEnums(pass *analysis.Pass) enums {
	pkgEnums := make(enums)

	// Gather enum types.
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if gen.Tok != token.TYPE {
				continue
			}
			for _, s := range gen.Specs {
				// Must be TypeSpec since we've filtered on token.TYPE.
				t, ok := s.(*ast.TypeSpec)
				obj := pass.TypesInfo.Defs[t.Name]
				if obj == nil {
					continue
				}

				named, ok := obj.Type().(*types.Named)
				if !ok {
					continue
				}
				basic, ok := named.Underlying().(*types.Basic)
				if !ok {
					continue
				}
				switch i := basic.Info(); {
				case i&types.IsInteger != 0:
					pkgEnums[named.Obj().Name()] = nil
				case i&types.IsFloat != 0:
					pkgEnums[named.Obj().Name()] = nil
				case i&types.IsString != 0:
					pkgEnums[named.Obj().Name()] = nil
				}
			}
		}
	}

	// Gather enum members.
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if gen.Tok != token.CONST && gen.Tok != token.VAR {
				continue
			}
			for _, s := range gen.Specs {
				// Must be ValueSpec since we've filtered on token.CONST, token.VAR.
				v := s.(*ast.ValueSpec)
				for _, name := range v.Names {
					obj := pass.TypesInfo.Defs[name]
					if obj == nil {
						continue
					}
					named, ok := obj.Type().(*types.Named)
					if !ok {
						continue
					}

					members, ok := pkgEnums[named.Obj().Name()]
					if !ok {
						continue
					}
					members = append(members, obj.Name())
					pkgEnums[named.Obj().Name()] = members
				}
			}
		}
	}

	// Delete member-less enum types.
	// We can't call these enums, since we can't be sure without
	// the existence of members. (The type may just be a named type,
	// for instance.)
	for k, v := range pkgEnums {
		if len(v) == 0 {
			delete(pkgEnums, k)
		}
	}

	return pkgEnums
}
