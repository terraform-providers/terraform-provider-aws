package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDmsReplicationSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDmsReplicationSubnetGroupCreate,
		Read:   resourceAwsDmsReplicationSubnetGroupRead,
		Update: resourceAwsDmsReplicationSubnetGroupUpdate,
		Delete: resourceAwsDmsReplicationSubnetGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"replication_subnet_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_subnet_group_description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_subnet_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDmsReplicationSubnetGroupId,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
			},
			"tags": tagsSchema(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDmsReplicationSubnetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	request := &dms.CreateReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier:  aws.String(d.Get("replication_subnet_group_id").(string)),
		ReplicationSubnetGroupDescription: aws.String(d.Get("replication_subnet_group_description").(string)),
		SubnetIds:                         expandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                              keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatabasemigrationserviceTags(),
	}

	log.Println("[DEBUG] DMS create replication subnet group:", request)

	_, err := conn.CreateReplicationSubnetGroup(request)
	if err != nil {
		return err
	}

	d.SetId(d.Get("replication_subnet_group_id").(string))
	return resourceAwsDmsReplicationSubnetGroupRead(d, meta)
}

func resourceAwsDmsReplicationSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	response, err := conn.DescribeReplicationSubnetGroups(&dms.DescribeReplicationSubnetGroupsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-subnet-group-id"),
				Values: []*string{aws.String(d.Id())}, // Must use d.Id() to work with import.
			},
		},
	})
	if err != nil {
		return err
	}
	if len(response.ReplicationSubnetGroups) == 0 {
		d.SetId("")
		return nil
	}

	// The AWS API for DMS subnet groups does not return the ARN which is required to
	// retrieve tags. This ARN can be built.
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "dms",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("subgrp:%s", d.Id()),
	}.String()
	d.Set("replication_subnet_group_arn", arn)

	err = resourceAwsDmsReplicationSubnetGroupSetState(d, response.ReplicationSubnetGroups[0])
	if err != nil {
		return err
	}

	tags, err := keyvaluetags.DatabasemigrationserviceListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for DMS Replication Subnet Group (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDmsReplicationSubnetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	// Updates to subnet groups are only valid when sending SubnetIds even if there are no
	// changes to SubnetIds.
	request := &dms.ModifyReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier: aws.String(d.Get("replication_subnet_group_id").(string)),
		SubnetIds:                        expandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if d.HasChange("replication_subnet_group_description") {
		request.ReplicationSubnetGroupDescription = aws.String(d.Get("replication_subnet_group_description").(string))
	}

	if d.HasChange("tags") {
		arn := d.Get("replication_subnet_group_arn").(string)
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatabasemigrationserviceUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating DMS Replication Subnet Group (%s) tags: %s", arn, err)
		}
	}

	log.Println("[DEBUG] DMS update replication subnet group:", request)

	_, err := conn.ModifyReplicationSubnetGroup(request)
	if err != nil {
		return err
	}

	return resourceAwsDmsReplicationSubnetGroupRead(d, meta)
}

func resourceAwsDmsReplicationSubnetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dmsconn

	request := &dms.DeleteReplicationSubnetGroupInput{
		ReplicationSubnetGroupIdentifier: aws.String(d.Get("replication_subnet_group_id").(string)),
	}

	log.Printf("[DEBUG] DMS delete replication subnet group: %#v", request)

	_, err := conn.DeleteReplicationSubnetGroup(request)
	return err
}

func resourceAwsDmsReplicationSubnetGroupSetState(d *schema.ResourceData, group *dms.ReplicationSubnetGroup) error {
	d.SetId(aws.StringValue(group.ReplicationSubnetGroupIdentifier))

	subnet_ids := []string{}
	for _, subnet := range group.Subnets {
		subnet_ids = append(subnet_ids, aws.StringValue(subnet.SubnetIdentifier))
	}

	d.Set("replication_subnet_group_description", group.ReplicationSubnetGroupDescription)
	d.Set("replication_subnet_group_id", group.ReplicationSubnetGroupIdentifier)
	d.Set("subnet_ids", subnet_ids)
	d.Set("vpc_id", group.VpcId)

	return nil
}
