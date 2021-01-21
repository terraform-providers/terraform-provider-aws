package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// See http://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions
// See https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-redshift.html
// See https://docs.amazonaws.cn/en_us/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions
var redshiftServiceAccountPerRegionMap = map[string]string{
	endpoints.AfSouth1RegionID:     "365689465814",
	endpoints.ApEast1RegionID:      "313564881002",
	endpoints.ApNortheast1RegionID: "404641285394",
	endpoints.ApNortheast2RegionID: "760740231472",
	"ap-northeast-3":               "090321488786", //lintignore:AWSAT003 // https://github.com/aws/aws-sdk-go/issues/1863
	endpoints.ApSouth1RegionID:     "865932855811",
	endpoints.ApSoutheast1RegionID: "361669875840",
	endpoints.ApSoutheast2RegionID: "762762565011",
	endpoints.CaCentral1RegionID:   "907379612154",
	endpoints.CnNorth1RegionID:     "111890595117",
	endpoints.CnNorthwest1RegionID: "660998842044",
	endpoints.EuCentral1RegionID:   "053454850223",
	endpoints.EuNorth1RegionID:     "729911121831",
	endpoints.EuSouth1RegionID:     "945612479654",
	endpoints.EuWest1RegionID:      "210876761215",
	endpoints.EuWest2RegionID:      "307160386991",
	endpoints.EuWest3RegionID:      "915173422425",
	endpoints.MeSouth1RegionID:     "013126148197",
	endpoints.SaEast1RegionID:      "075028567923",
	endpoints.UsEast1RegionID:      "193672423079",
	endpoints.UsEast2RegionID:      "391106570357",
	endpoints.UsGovEast1RegionID:   "665727464434",
	endpoints.UsGovWest1RegionID:   "665727464434",
	endpoints.UsWest1RegionID:      "262260360010",
	endpoints.UsWest2RegionID:      "902366379725",
}

func dataSourceAwsRedshiftServiceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRedshiftServiceAccountRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsRedshiftServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	region := meta.(*AWSClient).region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	if accid, ok := redshiftServiceAccountPerRegionMap[region]; ok {
		d.SetId(accid)
		arn := arn.ARN{
			Partition: meta.(*AWSClient).partition,
			Service:   "iam",
			AccountID: accid,
			Resource:  "user/logs",
		}.String()
		d.Set("arn", arn)

		return nil
	}

	return fmt.Errorf("Unknown region (%q)", region)
}
