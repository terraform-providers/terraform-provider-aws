package goheader

import "strings"

type AnalyzerOption interface {
	apply(*analyzer)
}

type applyAnalyzerOptionFunc func(*analyzer)

func (f applyAnalyzerOptionFunc) apply(a *analyzer) {
	f(a)
}

func WithValues(values map[string]Value) AnalyzerOption {
	return applyAnalyzerOptionFunc(func(a *analyzer) {
		a.values = make(map[string]Value)
		for k, v := range values {
			a.values[strings.ToLower(k)] = v
		}
	})
}

func WithTemplate(template string) AnalyzerOption {
	return applyAnalyzerOptionFunc(func(a *analyzer) {
		a.template = template
	})
}
