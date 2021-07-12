package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// alb/elb/nlb        - https://docs.aws.amazon.com/general/latest/gr/elb.html
// cloudfront         - https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html
// elastic_beanstalk  - https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html
// global_accelerator - https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html
// s3                 - https://docs.aws.amazon.com/general/latest/gr/s3.html
var hostedZoneIds = map[string]map[string]string{
	endpoints.AfSouth1RegionID: {
		"alb":                "Z268VQBMOI5EKX",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z1EI3BVKMKK4AM",
		"elb":                "Z268VQBMOI5EKX",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z203XCE67M25HM",
		"s3":                 "Z83WF9RJE8B12",
	},
	endpoints.ApEast1RegionID: {
		"alb":                "Z3DQVH9N71FHZ0",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "ZPWYUBWRU171A",
		"elb":                "Z3DQVH9N71FHZ0",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z12Y7K3UBGUAD1",
		"s3":                 "ZNB98KWMFR0R6",
	},
	endpoints.ApNortheast1RegionID: {
		"alb":                "Z14GRHDCWA56QT",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z1R25G3KIG2GBW",
		"elb":                "Z14GRHDCWA56QT",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z31USIVHYNEOWT",
		"s3":                 "Z2M4EHUR26P7ZW",
	},
	endpoints.ApNortheast2RegionID: {
		"alb":                "ZWKZPGTI48KDX",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z3JE5OI70TWKCP",
		"elb":                "ZWKZPGTI48KDX",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZIBE1TIR4HY56",
		"s3":                 "Z3W03O7B5YMIYP",
	},
	endpoints.ApNortheast3RegionID: {
		"alb":                "Z5LXEXXYW11ES",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "ZNE5GEY1TIAGY",
		"elb":                "Z5LXEXXYW11ES",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z1GWIQ4HH19I5X",
		"s3":                 "Z2YQB5RD63NC85",
	},
	endpoints.ApSouth1RegionID: {
		"alb":                "ZP97RAFLXTNZK",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z18NTBI3Y7N9TZ",
		"elb":                "ZP97RAFLXTNZK",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZVDDRBQ08TROA",
		"s3":                 "Z11RGJOFQNVJUP",
	},
	endpoints.ApSoutheast1RegionID: {
		"alb":                "Z1LMS91P8CMLE5",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z16FZ9L249IFLT",
		"elb":                "Z1LMS91P8CMLE5",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZKVM4W9LS7TM",
		"s3":                 "Z3O0J2DXBE1FTB",
	},
	endpoints.ApSoutheast2RegionID: {
		"alb":                "Z1GM3OXH4ZPM65",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z2PCDNR3VC2G1N",
		"elb":                "Z1GM3OXH4ZPM65",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZCT6FZBF4DROD",
		"s3":                 "Z1WCIGYICN2BYD",
	},
	endpoints.CaCentral1RegionID: {
		"alb":                "ZQSVJUPU6J1EY",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "ZJFCZL7SSZB5I",
		"elb":                "ZQSVJUPU6J1EY",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z2EPGBW3API2WT",
		"s3":                 "Z1QDHH18159H29",
	},
	endpoints.CnNorth1RegionID: {
		"alb":                "Z1GDH35T77C1KE",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "",
		"elb":                "Z1GDH35T77C1KE",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z3QFB96KMJ7ED6",
		"s3":                 "",
	},
	endpoints.CnNorthwest1RegionID: {
		"alb":                "ZM7IZAIOVVDZF",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "",
		"elb":                "ZM7IZAIOVVDZF",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZQEIKTCZ8352D",
		"s3":                 "Z282HJ1KT0DH03",
	},
	endpoints.EuCentral1RegionID: {
		"alb":                "Z215JYRZR1TBD5",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z1FRNW7UH4DEZJ",
		"elb":                "Z215JYRZR1TBD5",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z3F0SRJ5LGBH90",
		"s3":                 "Z21DNDUVLTQW6Q",
	},
	endpoints.EuNorth1RegionID: {
		"alb":                "Z23TAZ6LKFMNIO",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z23GO28BZ5AETM",
		"elb":                "Z23TAZ6LKFMNIO",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z1UDT6IFJ4EJM",
		"s3":                 "Z3BAZG2TWCNX0D",
	},
	endpoints.EuSouth1RegionID: {
		"alb":                "Z3ULH7SSC9OV64",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z10VDYYOA2JFKM",
		"elb":                "Z3ULH7SSC9OV64",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z23146JA1KNAFP",
		"s3":                 "Z30OZKI7KPW7MI",
	},
	endpoints.EuWest1RegionID: {
		"alb":                "Z32O12XQLNTSW2",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z2NYPWQ7DFZAZH",
		"elb":                "Z32O12XQLNTSW2",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z2IFOLAFXWLO4F",
		"s3":                 "Z1BKCTXD74EZPE",
	},
	endpoints.EuWest2RegionID: {
		"alb":                "ZHURV8PSTC4K8",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z1GKAAAUGATPF1",
		"elb":                "ZHURV8PSTC4K8",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZD4D7Y8KGAS4G",
		"s3":                 "Z3GKZC51ZF0DB4",
	},
	endpoints.EuWest3RegionID: {
		"alb":                "Z3Q77PNBQS71R4",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z5WN6GAYWG5OB",
		"elb":                "Z3Q77PNBQS71R4",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z1CMS0P5QUZ6D5",
		"s3":                 "Z3R1K369G5AVDG",
	},
	endpoints.MeSouth1RegionID: {
		"alb":                "ZS929ML54UICD",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z2BBTEKR2I36N2",
		"elb":                "ZS929ML54UICD",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z3QSRYVP46NYYV",
		"s3":                 "Z1MPMWCPA7YB62",
	},
	endpoints.SaEast1RegionID: {
		"alb":                "Z2P70J7HTTTPLU",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z10X7K2B4QSOFV",
		"elb":                "Z2P70J7HTTTPLU",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZTK26PT1VY4CU",
		"s3":                 "Z7KQH4QJS55SO",
	},
	endpoints.UsEast1RegionID: {
		"alb":                "Z35SXDOTRQ7X7K",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z117KPS5GTRQ2G",
		"elb":                "Z35SXDOTRQ7X7K",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z26RNL4JYFTOTI",
		"s3":                 "Z3AQBSTGFYJSTF",
	},
	endpoints.UsEast2RegionID: {
		"alb":                "Z3AADJGX6KTTL2",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z14LCN19Q5QHIC",
		"elb":                "Z3AADJGX6KTTL2",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZLMOA37VPKANP",
		"s3":                 "Z2O1EMRO9K5GLX",
	},
	endpoints.UsGovEast1RegionID: {
		"alb":                "Z166TLBEWOO7G0",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z35TSARG0EJ4VU",
		"elb":                "Z166TLBEWOO7G0",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z1ZSMQQ6Q24QQ8",
		"s3":                 "Z2NIFVYYW2VKV1",
	},
	endpoints.UsGovWest1RegionID: {
		"alb":                "Z33AYJ8TM3BH4J",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z4KAURWC4UUUG",
		"elb":                "Z33AYJ8TM3BH4J",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "ZMG1MZ2THAWF1",
		"s3":                 "Z31GFT0UA1I2HV",
	},
	endpoints.UsWest1RegionID: {
		"alb":                "Z368ELLRRE2KJ0",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z1LQECGX5PH1X",
		"elb":                "Z368ELLRRE2KJ0",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z24FKFUX50B4VW",
		"s3":                 "Z2F56UZL2M1ACD",
	},
	endpoints.UsWest2RegionID: {
		"alb":                "Z1H1FL5HABSF5",
		"cloudfront":         "Z2FDTNDATAQYW2",
		"elastic_beanstalk":  "Z38NKT9BP95V3O",
		"elb":                "Z1H1FL5HABSF5",
		"global_accelerator": "Z2BJ6XQ5FK7U4H",
		"nlb":                "Z18D5FSROUN65G",
		"s3":                 "Z3BJ6K6RIION7M",
	},
}

func dataSourceAwsHostedZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsHostedZonesRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"alb": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elastic_beanstalk": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elb": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_accelerator": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nlb": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsHostedZonesRead(d *schema.ResourceData, meta interface{}) error {
	region := meta.(*AWSClient).region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	ids, ok := hostedZoneIds[region]

	if !ok {
		return fmt.Errorf("Unsupported region: %s", region)
	}

	d.SetId(region)
	d.Set("region", region)
	d.Set("alb", ids["alb"])
	d.Set("cloudfront", ids["cloudfront"])
	d.Set("elastic_beanstalk", ids["elastic_beanstalk"])
	d.Set("elb", ids["elb"])
	d.Set("global_accelerator", ids["global_accelerator"])
	d.Set("nlb", ids["nlb"])
	d.Set("s3", ids["s3"])

	return nil
}
