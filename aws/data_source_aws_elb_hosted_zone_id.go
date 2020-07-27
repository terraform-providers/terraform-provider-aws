package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// See https://docs.aws.amazon.com/general/latest/gr/rande.html#elb_region
// See https://docs.amazonaws.cn/en_us/general/latest/gr/rande.html#elb_region
// First index is Application/Classic LB id's, second index is NLB id's
var elbHostedZoneIdPerRegionMap = map[string][]string{
	"af-south-1":     {"Z268VQBMOI5EKX", "Z203XCE67M25HM"},
	"ap-east-1":      {"Z3DQVH9N71FHZ0", "Z12Y7K3UBGUAD1"},
	"ap-northeast-1": {"Z14GRHDCWA56QT", "Z31USIVHYNEOWT"},
	"ap-northeast-2": {"ZWKZPGTI48KDX", "ZIBE1TIR4HY56"},
	"ap-northeast-3": {"Z5LXEXXYW11ES", "Z1GWIQ4HH19I5X"},
	"ap-south-1":     {"ZP97RAFLXTNZK", "ZVDDRBQ08TROA"},
	"ap-southeast-1": {"Z1LMS91P8CMLE5", "ZKVM4W9LS7TM"},
	"ap-southeast-2": {"Z1GM3OXH4ZPM65", "ZCT6FZBF4DROD"},
	"ca-central-1":   {"ZQSVJUPU6J1EY", "Z2EPGBW3API2WT"},
	"cn-north-1":     {"Z1GDH35T77C1KE", "Z3QFB96KMJ7ED6"},
	"cn-northwest-1": {"ZM7IZAIOVVDZF", "ZQEIKTCZ8352D"},
	"eu-central-1":   {"Z215JYRZR1TBD5", "Z3F0SRJ5LGBH90"},
	"eu-north-1":     {"Z23TAZ6LKFMNIO", "Z1UDT6IFJ4EJM"},
	"eu-south-1":     {"Z3ULH7SSC9OV64", "Z23146JA1KNAFP"},
	"eu-west-1":      {"Z32O12XQLNTSW2", "Z2IFOLAFXWLO4F"},
	"eu-west-2":      {"ZHURV8PSTC4K8", "ZD4D7Y8KGAS4G"},
	"eu-west-3":      {"Z3Q77PNBQS71R4", "Z1CMS0P5QUZ6D5"},
	"me-south-1":     {"ZS929ML54UICD", "Z3QSRYVP46NYYV"},
	"sa-east-1":      {"Z2P70J7HTTTPLU", "ZTK26PT1VY4CU"},
	"us-east-1":      {"Z35SXDOTRQ7X7K", "Z26RNL4JYFTOTI"},
	"us-east-2":      {"Z3AADJGX6KTTL2", "ZLMOA37VPKANP"},
	"us-gov-east-1":  {"Z166TLBEWOO7G0", "Z1ZSMQQ6Q24QQ8"},
	"us-gov-west-1":  {"Z33AYJ8TM3BH4J", "ZMG1MZ2THAWF1"},
	"us-west-1":      {"Z368ELLRRE2KJ0", "Z24FKFUX50B4VW"},
	"us-west-2":      {"Z1H1FL5HABSF5", "Z18D5FSROUN65G"},
}

var validElbTypes = []string{
	// elbv2 and elb don't have an option for classic
	elbv2.LoadBalancerTypeEnumApplication,
	"classic",
	elbv2.LoadBalancerTypeEnumNetwork,
}

func dataSourceAwsElbHostedZoneId() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElbHostedZoneIdRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"elb_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validElbTypes, false),
			},
		},
	}
}

func dataSourceAwsElbHostedZoneIdRead(d *schema.ResourceData, meta interface{}) error {
	region := meta.(*AWSClient).region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}
	var lbTypeIndex int
	if v, ok := d.GetOk("elb_type"); ok {
		if v.(string) == "classic" || v.(string) == "application" {
			lbTypeIndex = 0
		} else if v.(string) == "network" {
			lbTypeIndex = 1
		}
	} else {
		lbTypeIndex = 0
	}
	if zoneId, ok := elbHostedZoneIdPerRegionMap[region]; ok {
		d.SetId(zoneId[lbTypeIndex])
		return nil
	}

	return fmt.Errorf("Unknown region (%q)", region)
}
