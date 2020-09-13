// Package exhaustive provides an analyzer that helps ensure enum switch statements
// are exhaustive. The analyzer also provides fixes to make the offending switch
// statements exhaustive (see "Fixes" section).
//
// See "cmd/exhaustive" subpackage for the related command line program.
//
// Definition of enum
//
// The language spec does not provide an explicit definition for enums.
// For the purpose of this program, an enum type is a package-level named type
// whose underlying type is an integer (includes byte and rune), a float, or
// a string type. An enum type must have associated with it one or more
// package-level variables of the named type in the package. These variables
// constitute the enum's members.
//
// In the code snippet below, Biome is an enum type with 3 members.
//
//   type Biome int
//
//   const (
//       Tundra Biome = iota
//       Savanna
//       Desert
//   )
//
// Switch statement exhaustiveness
//
// An enum switch statement is exhaustive if it has cases for each of the enum's members.
//
// For an enum type defined in the same package as the switch statement, both
// exported and unexported enum members must be present in order to consider
// the switch exhaustive. On the other hand, for an enum type defined
// in an external package it is sufficient for just exported enum members
// to be present in order to consider the switch exhaustive.
//
// Flags
//
// The analyzer accepts a boolean flag: -default-signifies-exhaustive.
// The flag, if set, indicates to the analyzer that switch statements
// are to be considered exhaustive as long as a 'default' case is present, even
// if all enum members aren't listed in the switch statements cases.
//
// The other relevant flag is the -fix flag.
//
// Fixes
//
// The analyzer suggests fixes for a switch statement if it is not exhaustive
// and does not have a 'default' case. The suggested fix always adds a single
// case clause for the missing enum members.
//
//   case missingA, missingB, missingC:
//       panic(fmt.Sprintf("unhandled value: %v", v))
//
// where v is the expression in the switch statement's tag (in other words, the
// value being switched upon). If the switch statement's tag is a function or a
// method call the analyzer does not suggest a fix, as reusing the call expression
// in the panic/fmt.Sprintf call could be mutative.
//
// The rationale for the fix using panic is that it might be better to fail loudly on
// existing unhandled or impossible cases than to let them slip by quietly unnoticed.
// An even better fix may, of course, be to manually inspect the sites reported
// by the package and handle the missing cases if necessary.
//
// Imports will be adjusted automatically to account for the "fmt" dependency.
//
// Skip analysis of specific switch statements
//
// If the following directive comment:
//
//   //exhaustive:ignore
//
// is associated with a switch statement, the analyzer skips
// checking of the switch statement and no diagnostics are reported.
package exhaustive

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	// DefaultSignifiesExhaustiveFlag is a flag name used by the analyzer. It
	// is exported for use by analyzer driver programs.
	DefaultSignifiesExhaustiveFlag = "default-signifies-exhaustive"
)

var (
	fCheckMaps                  bool
	fDefaultSignifiesExhaustive bool
)

func init() {
	Analyzer.Flags.BoolVar(&fCheckMaps, "maps", false, "check key exhaustiveness for map literals of enum key type, in addition to checking switch statements")
	Analyzer.Flags.BoolVar(&fDefaultSignifiesExhaustive, DefaultSignifiesExhaustiveFlag, false, "indicates that switch statements are to be considered exhaustive if a 'default' case is present, even if all enum members aren't listed in the switch")
}

var Analyzer = &analysis.Analyzer{
	Name:      "exhaustive",
	Doc:       "check exhaustiveness of enum switch statements",
	Run:       run,
	Requires:  []*analysis.Analyzer{inspect.Analyzer},
	FactTypes: []analysis.Fact{&enumsFact{}},
}

// IgnoreDirectivePrefix is used to exclude checking of specific switch statements.
// See https://godoc.org/github.com/nishanths/exhaustive#hdr-Skip_analysis_of_specific_switch_statements
// for details.
const IgnoreDirectivePrefix = "//exhaustive:ignore"

func containsIgnoreDirective(comments []*ast.Comment) bool {
	for _, c := range comments {
		if strings.HasPrefix(c.Text, IgnoreDirectivePrefix) {
			return true
		}
	}
	return false
}

type enumsFact struct {
	Entries enums
}

var _ analysis.Fact = (*enumsFact)(nil)

func (e *enumsFact) AFact() {}

func (e *enumsFact) String() string {
	// sort for stability (required for testing)
	var sortedKeys []string
	for k := range e.Entries {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var buf strings.Builder
	for i, k := range sortedKeys {
		v := e.Entries[k]
		buf.WriteString(k)
		buf.WriteString(":")
		for j, vv := range v {
			buf.WriteString(vv)
			// add comma separator between each enum member in an enum type
			if j != len(v)-1 {
				buf.WriteString(",")
			}
		}
		// add semicolon separator between each enum type
		if i != len(sortedKeys)-1 {
			buf.WriteString("; ")
		}
	}
	return buf.String()
}

func run(pass *analysis.Pass) (interface{}, error) {
	e := findEnums(pass)
	if len(e) != 0 {
		pass.ExportPackageFact(&enumsFact{Entries: e})
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	comments := make(map[*ast.File]ast.CommentMap) // CommentMap per package file, lazily populated by reference

	checkSwitchStatements(pass, inspect, comments)
	if fCheckMaps {
		checkMapLiterals(pass, inspect, comments)
	}
	return nil, nil
}

func enumTypeName(e *types.Named, samePkg bool) string {
	if samePkg {
		return e.Obj().Name()
	}
	return e.Obj().Pkg().Name() + "." + e.Obj().Name()
}
