package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDefaultDbSubnetGroup() *schema.Resource {
	// reuse aws_db_subnet_group schema, and methods for READ
	defDbSubnetGroup := resourceAwsDbSubnetGroup()
	defDbSubnetGroup.Create = resourceAwsDefaultDbSubnetGroupCreate
	defDbSubnetGroup.Delete = resourceAwsDefaultDbSubnetGroupDelete
	defDbSubnetGroup.Update = resourceAwsDefaultDbSubnetGroupUpdate

	// name is a computed value for Default DB Subnet Group
	defDbSubnetGroup.Schema["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	delete(defDbSubnetGroup.Schema, "name_prefix")

	return defDbSubnetGroup
}

func resourceAwsDefaultDbSubnetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	rdsconn := meta.(*AWSClient).rdsconn

	describeOpts := rds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String("default"),
	}

	describeResp, err := rdsconn.DescribeDBSubnetGroups(&describeOpts)
	if err != nil {
		return err
	}

	if describeResp.DBSubnetGroups == nil || len(describeResp.DBSubnetGroups) == 0 {
		return fmt.Errorf("No default DB Subnet Group found in this region.")
	}

	subnetGroup := describeResp.DBSubnetGroups[0]
	d.SetId(aws.StringValue(subnetGroup.DBSubnetGroupName))
	d.Set("arn", subnetGroup.DBSubnetGroupArn)

	return resourceAwsDefaultDbSubnetGroupUpdate(d, meta)
}

func resourceAwsDefaultDbSubnetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	d.Partial(true)

	if d.HasChange("subnet_ids") {
		_, n := d.GetChange("subnet_ids")
		if n == nil {
			n = new(schema.Set)
		}

		ns := n.(*schema.Set)

		var sIds []*string
		for _, s := range ns.List() {
			sIds = append(sIds, aws.String(s.(string)))
		}

		_, err := conn.ModifyDBSubnetGroup(&rds.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                sIds,
		})

		if err != nil {
			return err
		}
	}

	if d.HasChange("description") && !d.HasChange("subnet_ids") {
		subnetIdsSet := d.Get("subnet_ids").(*schema.Set)
		sIds := make([]*string, subnetIdsSet.Len())
		for i, subnetId := range subnetIdsSet.List() {
			sIds[i] = aws.String(subnetId.(string))
		}

		_, err := conn.ModifyDBSubnetGroup(&rds.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                sIds,
		})

		if err != nil {
			return err
		}
	}

	arn := d.Get("arn").(string)
	if err := setTagsRDS(conn, d, arn); err != nil {
		return err
	} else {
		d.SetPartial("tags")
	}

	d.Partial(false)

	return resourceAwsDbSubnetGroupRead(d, meta)
}

func resourceAwsDefaultDbSubnetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default DB Subnet Group. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
