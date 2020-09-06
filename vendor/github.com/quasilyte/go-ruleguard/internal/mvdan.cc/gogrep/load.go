// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package gogrep

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func (m *matcher) load(wd string, args ...string) ([]*packages.Package, error) {
	mode := packages.NeedName | packages.NeedImports | packages.NeedSyntax |
		packages.NeedTypes | packages.NeedTypesInfo
	if m.recursive { // need the syntax trees for the dependencies too
		mode |= packages.NeedDeps
	}
	cfg := &packages.Config{
		Mode:  mode,
		Dir:   wd,
		Fset:  m.fset,
		Tests: m.tests,
	}
	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		return nil, err
	}
	jointErr := ""
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			jointErr += err.Error() + "\n"
		}
	})
	if jointErr != "" {
		return nil, fmt.Errorf("%s", jointErr)
	}

	// Make a sorted list of the packages, including transitive dependencies
	// if recurse is true.
	byPath := make(map[string]*packages.Package)
	var addDeps func(*packages.Package)
	addDeps = func(pkg *packages.Package) {
		if strings.HasSuffix(pkg.PkgPath, ".test") {
			// don't add recursive test deps
			return
		}
		for _, imp := range pkg.Imports {
			if _, ok := byPath[imp.PkgPath]; ok {
				continue // seen; avoid recursive call
			}
			byPath[imp.PkgPath] = imp
			addDeps(imp)
		}
	}
	for _, pkg := range pkgs {
		byPath[pkg.PkgPath] = pkg
		if m.recursive {
			// add all dependencies once
			addDeps(pkg)
		}
	}
	pkgs = pkgs[:0]
	for _, pkg := range byPath {
		pkgs = append(pkgs, pkg)
	}
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].PkgPath < pkgs[j].PkgPath
	})
	return pkgs, nil
}
