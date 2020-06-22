package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsEfsMountTarget() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEfsMountTargetRead,

		Schema: map[string]*schema.Schema{
			"mount_target_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_target_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEfsMountTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	describeEfsOpts := &efs.DescribeMountTargetsInput{
		MountTargetId: aws.String(d.Get("mount_target_id").(string)),
	}

	log.Printf("[DEBUG] Reading EFS Mount Target: %s", describeEfsOpts)
	resp, err := conn.DescribeMountTargets(describeEfsOpts)
	if err != nil {
		return fmt.Errorf("Error retrieving EFS Mount Target: %s", err)
	}
	if len(resp.MountTargets) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(resp.MountTargets))
	}

	mt := resp.MountTargets[0]

	log.Printf("[DEBUG] Found EFS mount target: %#v", mt)

	d.SetId(aws.StringValue(mt.MountTargetId))

	fsARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(mt.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()

	d.Set("file_system_arn", fsARN)
	d.Set("file_system_id", mt.FileSystemId)
	d.Set("ip_address", mt.IpAddress)
	d.Set("subnet_id", mt.SubnetId)
	d.Set("network_interface_id", mt.NetworkInterfaceId)
	d.Set("availability_zone_name", mt.AvailabilityZoneName)
	d.Set("availability_zone_id", mt.AvailabilityZoneId)
	d.Set("owner_id", mt.OwnerId)

	sgResp, err := conn.DescribeMountTargetSecurityGroups(&efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}
	err = d.Set("security_groups", flattenStringSet(sgResp.SecurityGroups))
	if err != nil {
		return err
	}

	d.Set("dns_name", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(mt.FileSystemId))))
	d.Set("mount_target_dns_name", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.%s.efs", aws.StringValue(mt.AvailabilityZoneName), aws.StringValue(mt.FileSystemId))))

	return nil
}
