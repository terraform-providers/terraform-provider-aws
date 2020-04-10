package passes

import (
	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSAT001"
	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSR001"
	"golang.org/x/tools/go/analysis"
)

var AllChecks = []*analysis.Analyzer{
	AWSAT001.Analyzer,
	AWSR001.Analyzer,
}
