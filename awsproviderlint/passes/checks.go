package passes

import (
	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSAT001"
	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSR001"
	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSR002"
	"golang.org/x/tools/go/analysis"
)

var AllChecks = []*analysis.Analyzer{
	AWSAT001.Analyzer,
	AWSR001.Analyzer,
	AWSR002.Analyzer,
}
