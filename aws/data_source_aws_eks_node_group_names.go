package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsEksNodeGroupNames() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksNodeGroupNamesRead,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsEksNodeGroupNamesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	clusterName := d.Get("cluster_name").(string)

	input := &eks.ListNodegroupsInput{
		ClusterName: aws.String(clusterName),
	}

	log.Printf("[DEBUG] Reading EKS Node Groups: %s", input)
	output, err := conn.ListNodegroups(input)
	if err != nil {
		return err
	}

	if output == nil || output.Nodegroups == nil {
		return fmt.Errorf("EKS Node Groups on cluster (%s) not found", clusterName)
	}

	nodeGroups := output.Nodegroups

	d.SetId(clusterName)
	d.Set("cluster_name", clusterName)
	d.Set("names", nodeGroups)

	return nil
}
