// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	outputName = "list_pages_gen.go"
)

var (
	functionNames = flag.String("function", "", "comma-separated list of API List functions; required")
	paginatorName = flag.String("paginator", "NextToken", "name of the pagination token field")
	packageName   = flag.String("package", "", "override package name for generated code")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] -function <function-name>[,<function-name>] <source-package>\n\n")
	fmt.Fprintf(os.Stderr, "\tDestination package is read from the environment variable $GOPACKAGE by default. Override it with the flag -package.\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	destinationPackage := os.Getenv("GOPACKAGE")
	if *packageName != "" {
		destinationPackage = *packageName
	}
	args := flag.Args()
	if len(*functionNames) == 0 || len(args) == 0 || destinationPackage == "" {
		flag.Usage()
		os.Exit(2)
	}
	sourcePackage := args[0]
	functions := strings.Split(*functionNames, ",")
	sort.Strings(functions)

	g := Generator{
		paginator: *paginatorName,
		tmpl:      template.Must(template.New("function").Parse(functionTemplate)),
	}
	g.parsePackage(sourcePackage)

	g.printHeader(HeaderInfo{
		Parameters:         strings.Join(os.Args[1:], " "),
		DestinationPackage: destinationPackage,
		SourcePackage:      sourcePackage,
	})

	for _, functionName := range functions {
		g.generateFunction(functionName)
	}

	src := g.format()

	err := ioutil.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("error writing output: %s", err)
	}
}

type HeaderInfo struct {
	Parameters         string
	DestinationPackage string
	SourcePackage      string
}

type Generator struct {
	buf       bytes.Buffer
	pkg       *Package
	tmpl      *template.Template
	paginator string
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

type PackageFile struct {
	file *ast.File
}

type Package struct {
	name  string
	files []*PackageFile
}

func (g *Generator) printHeader(headerInfo HeaderInfo) {
	header := template.Must(template.New("header").Parse(headerTemplate))
	err := header.Execute(&g.buf, headerInfo)
	if err != nil {
		log.Fatalf("error writing header: %s", err)
	}
}

func (g *Generator) parsePackage(sourcePackage string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, sourcePackage)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		files: make([]*PackageFile, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &PackageFile{
			file: file,
		}
	}
}

type FuncSpec struct {
	Name       string
	RecvType   string
	ParamType  string
	ResultType string
	Paginator  string
}

func (g *Generator) generateFunction(functionName string) {
	var function *ast.FuncDecl

	// TODO: check if a Pages() function has been defined
	for _, file := range g.pkg.files {
		if file.file != nil {
			for _, decl := range file.file.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					if funcDecl.Name.Name == functionName {
						function = funcDecl
						break
					}
				}
			}
			if function != nil {
				break
			}
		}
	}

	if function == nil {
		log.Fatalf("function \"%s\" not found", functionName)
	}

	funcSpec := FuncSpec{
		Name:       function.Name.Name,
		RecvType:   g.expandTypeField(function.Recv),
		ParamType:  g.expandTypeField(function.Type.Params),  // Assumes there is a single input parameter
		ResultType: g.expandTypeField(function.Type.Results), // Assumes we can take the first return parameter
		Paginator:  g.paginator,
	}

	err := g.tmpl.Execute(&g.buf, funcSpec)
	if err != nil {
		log.Fatalf("error writing function \"%s\": %s", functionName, err)
	}
}

func (g *Generator) expandTypeField(field *ast.FieldList) string {
	typeValue := field.List[0].Type
	if star, ok := typeValue.(*ast.StarExpr); ok {
		return fmt.Sprintf("*%s", g.expandTypeExpr(star.X))
	}

	log.Fatalf("Unexpected type expression: (%[1]T) %[1]v", typeValue)
	return ""
}

func (g *Generator) expandTypeExpr(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return fmt.Sprintf("%s.%s", g.pkg.name, ident.Name)
	}

	log.Fatalf("Unexpected expression: (%[1]T) %[1]v", expr)
	return ""
}

const headerTemplate = `// Code generated by "aws/internal/generators/listpages/main.go {{ .Parameters }}"; DO NOT EDIT.

package {{ .DestinationPackage }}

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"{{ .SourcePackage }}"
)
`

const functionTemplate = `

func {{ .Name }}Pages(conn {{ .RecvType }}, input {{ .ParamType }}, fn func({{ .ResultType }}, bool) bool) error {
	return {{ .Name }}PagesWithContext(context.Background(), conn, input, fn)
}

func {{ .Name }}PagesWithContext(ctx context.Context, conn {{ .RecvType }}, input {{ .ParamType }}, fn func({{ .ResultType }}, bool) bool) error {
	for {
		output, err := conn.{{ .Name }}WithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.{{ .Paginator }}) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.{{ .Paginator }} = output.{{ .Paginator }}
	}
	return nil
}
`

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}
