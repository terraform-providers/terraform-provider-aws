package AT012

import (
	"flag"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testaccfuncdecl"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for test files containing multiple acceptance test function name prefixes

The AT012 analyzer reports likely incorrect uses of multiple TestAcc function
name prefixes up to the conventional underscore (_) prefix separator within
the same file. Typically, Terraform acceptance tests should use the same naming
prefix within one test file so testers can easily run all acceptance tests for
the file and not miss associated tests.

Optional parameters:
  - ignored-filenames Comma-separated list of file names to ignore, defaults to none.`

const (
	acceptanceTestNameSeparator = "_"

	analyzerName = "AT012"
)

var (
	ignoredFilenames string
)

var Analyzer = &analysis.Analyzer{
	Name:  analyzerName,
	Doc:   Doc,
	Flags: parseFlags(),
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testaccfuncdecl.Analyzer,
	},
	Run: run,
}

func isFilenameIgnored(fileName string, fileNameList string) bool {
	prefixes := strings.Split(fileNameList, ",")

	for _, prefix := range prefixes {
		if strings.HasPrefix(fileName, prefix) {
			return true
		}
	}
	return false
}

func parseFlags() flag.FlagSet {
	var flags = flag.NewFlagSet(analyzerName, flag.ExitOnError)
	flags.StringVar(&ignoredFilenames, "ignored-filenames", "", "Comma-separated list of file names to ignore")
	return *flags
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	funcDecls := pass.ResultOf[testaccfuncdecl.Analyzer].([]*ast.FuncDecl)

	fileFuncDecls := make(map[*token.File][]*ast.FuncDecl)

	for _, funcDecl := range funcDecls {
		file := pass.Fset.File(funcDecl.Pos())
		fileName := filepath.Base(file.Name())

		if ignoredFilenames != "" && isFilenameIgnored(fileName, ignoredFilenames) {
			continue
		}

		if ignorer.ShouldIgnore(analyzerName, funcDecl) {
			continue
		}

		fileFuncDecls[file] = append(fileFuncDecls[file], funcDecl)
	}

	for file, funcDecls := range fileFuncDecls {
		// Map to simplify checking
		funcNamePrefixes := make(map[string]struct{})

		for _, funcDecl := range funcDecls {
			funcName := funcDecl.Name.Name

			funcNamePrefixParts := strings.SplitN(funcName, acceptanceTestNameSeparator, 2)

			// Ensure function name includes separator
			if len(funcNamePrefixParts) != 2 || funcNamePrefixParts[0] == "" || funcNamePrefixParts[1] == "" {
				continue
			}

			funcNamePrefix := funcNamePrefixParts[0]

			funcNamePrefixes[funcNamePrefix] = struct{}{}
		}

		if len(funcNamePrefixes) <= 1 {
			continue
		}

		// Easier to print map keys as slice
		namePrefixes := make([]string, 0, len(funcNamePrefixes))
		for k := range funcNamePrefixes {
			namePrefixes = append(namePrefixes, k)
		}

		pass.Reportf(file.Pos(0), "%s: file contains multiple acceptance test name prefixes: %v", analyzerName, namePrefixes)
	}

	return nil, nil
}
