package goheader

import (
	"fmt"
	"go/ast"
	"strings"
)

type Analyzer interface {
	Analyze(file *ast.File) Issue
}

type analyzer struct {
	values   map[string]Value
	template string
}

func (a *analyzer) Analyze(file *ast.File) Issue {
	if a.template == "" {
		return NewIssue("Missed template for check")
	}
	var header string
	if len(file.Comments) > 0 && file.Comments[0].Pos() < file.Package {
		if strings.HasPrefix(file.Comments[0].List[0].Text, "/*") {
			header = (&ast.CommentGroup{List: []*ast.Comment{file.Comments[0].List[0]}}).Text()
		} else {
			header = file.Comments[0].Text()
		}
	}
	header = strings.TrimSpace(header)
	if header == "" {
		return NewIssue("Missed header for check")
	}
	s := NewReader(header)
	t := NewReader(a.template)
	for !s.Done() && !t.Done() {
		templateCh := t.Peek()
		if templateCh == '{' {
			name := a.readField(t)
			if a.values[name] == nil {
				return NewIssue(fmt.Sprintf("Template has unknown value: %v", name))
			}
			if i := a.values[name].Read(s); i != nil {
				return i
			}
			continue
		}
		sourceCh := s.Peek()
		if sourceCh != templateCh {
			l := s.Location()
			notNextLine := func(r rune) bool {
				return r != '\n'
			}
			actual := s.ReadWhile(notNextLine)
			expected := t.ReadWhile(notNextLine)
			return NewIssueWithLocation(fmt.Sprintf("Actual: %v\nExpected:%v", actual, expected), l)
		}
		s.Next()
		t.Next()
	}
	if !s.Done() {
		l := s.Location()
		return NewIssueWithLocation(fmt.Sprintf("Unexpected string: %v", s.Finish()), l)
	}
	if !t.Done() {
		l := s.Location()
		return NewIssueWithLocation(fmt.Sprintf("Missed string: %v", t.Finish()), l)
	}
	return nil
}

func (a *analyzer) readField(reader Reader) string {
	_ = reader.Next()
	_ = reader.Next()

	r := reader.ReadWhile(func(r rune) bool {
		return r != '}'
	})

	_ = reader.Next()
	_ = reader.Next()

	return strings.ToLower(strings.TrimSpace(r))
}

func New(options ...AnalyzerOption) Analyzer {
	a := &analyzer{}
	for _, o := range options {
		o.apply(a)
	}
	for _, v := range a.values {
		err := v.Calculate(a.values)
		if err != nil {
			panic(err.Error())
		}
	}
	return a
}
