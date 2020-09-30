package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsDmsReplicationInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDmsReplicationInstanceRead,

		Schema: map[string]*schema.Schema{
			"replication_instance_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"replication_instance_private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"replication_instance_public_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsDmsReplicationInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).dmsconn

	filters, filtersOk := d.GetOk("filter")

	// Build up search parameters
	params := &dms.DescribeReplicationInstancesInput{}
	if filtersOk {
		params.Filters = buildAwsDmsDataSourceFilters(filters.(*schema.Set))
	}

	log.Printf("[DEBUG] Reading Replication Instances")
	res, err := client.DescribeReplicationInstances(params)

	if err != nil {
		return fmt.Errorf("Error getting Replication Instances: %v", err)
	}

	var instance *dms.ReplicationInstance
	if len(res.ReplicationInstances) < 1 {
		return nil
	}

	if len(res.ReplicationInstances) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more " +
			"specific search criteria.")
	} else {
		instance = res.ReplicationInstances[0]
	}

	log.Printf("[DEBUG] Received Replication Instances: %s", res)

	d.SetId(time.Now().UTC().String())
	d.Set("replication_instance_arn", instance.ReplicationInstanceArn)
	d.Set("replication_instance_private_ips", instance.ReplicationInstancePrivateIpAddresses)
	d.Set("replication_instance_public_ips", instance.ReplicationInstancePublicIpAddresses)

	return nil
}

func buildAwsDmsDataSourceFilters(set *schema.Set) []*dms.Filter {
	var dms_filters []*dms.Filter
	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}
		dms_filters = append(dms_filters, &dms.Filter{
			Name:   aws.String(m["name"].(string)),
			Values: filterValues,
		})
	}
	return dms_filters
}
