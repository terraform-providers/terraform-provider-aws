package xpasses

import (
	"github.com/bflad/tfproviderlint/xpasses/XAT001"
	"github.com/bflad/tfproviderlint/xpasses/XR001"
	"github.com/bflad/tfproviderlint/xpasses/XR002"
	"github.com/bflad/tfproviderlint/xpasses/XR003"
	"github.com/bflad/tfproviderlint/xpasses/XR004"
	"github.com/bflad/tfproviderlint/xpasses/XR005"
	"github.com/bflad/tfproviderlint/xpasses/XR006"
	"github.com/bflad/tfproviderlint/xpasses/XR007"
	"github.com/bflad/tfproviderlint/xpasses/XR008"
	"github.com/bflad/tfproviderlint/xpasses/XS001"
	"github.com/bflad/tfproviderlint/xpasses/XS002"
	"golang.org/x/tools/go/analysis"
)

// AllChecks contains all Analyzers that report issues
// This can be consumed via multichecker.Main(xpasses.AllChecks...) or by
// combining these Analyzers with additional custom Analyzers
var AllChecks = []*analysis.Analyzer{
	XAT001.Analyzer,
	XR001.Analyzer,
	XR002.Analyzer,
	XR003.Analyzer,
	XR004.Analyzer,
	XR005.Analyzer,
	XR006.Analyzer,
	XR007.Analyzer,
	XR008.Analyzer,
	XS001.Analyzer,
	XS002.Analyzer,
}
