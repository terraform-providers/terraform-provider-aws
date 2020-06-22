package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsPlacementGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPlacementGroupCreate,
		Read:   resourceAwsPlacementGroupRead,
		Update: resourceAwsPlacementGroupUpdate,
		Delete: resourceAwsPlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"strategy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.PlacementStrategyCluster,
					ec2.PlacementStrategyPartition,
					ec2.PlacementStrategySpread,
				}, false),
			},
			"placement_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsPlacementGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	name := d.Get("name").(string)
	input := ec2.CreatePlacementGroupInput{
		GroupName:         aws.String(name),
		Strategy:          aws.String(d.Get("strategy").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypePlacementGroup),
	}
	log.Printf("[DEBUG] Creating EC2 Placement group: %s", input)
	_, err := conn.CreatePlacementGroup(&input)
	if err != nil {
		return err
	}

	wait := resource.StateChangeConf{
		Pending:    []string{ec2.PlacementGroupStatePending},
		Target:     []string{ec2.PlacementGroupStateAvailable},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
				GroupNames: []*string{aws.String(name)},
			})

			if err != nil {
				// Fix timing issue where describe is called prior to
				// create being effectively processed by AWS
				if isAWSErr(err, "InvalidPlacementGroup.Unknown", "") {
					return out, "pending", nil
				}
				return out, "", err
			}

			if len(out.PlacementGroups) == 0 {
				return out, "", fmt.Errorf("Placement group not found (%q)", name)
			}
			pg := out.PlacementGroups[0]

			return out, *pg.State, nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] EC2 Placement group created: %q", name)

	d.SetId(name)

	return resourceAwsPlacementGroupRead(d, meta)
}

func resourceAwsPlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := ec2.DescribePlacementGroupsInput{
		GroupNames: []*string{aws.String(d.Id())},
	}
	out, err := conn.DescribePlacementGroups(&input)
	if err != nil {
		if isAWSErr(err, "InvalidPlacementGroup.Unknown", "") {
			log.Printf("[WARN] Placement Group %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}
	pg := out.PlacementGroups[0]

	log.Printf("[DEBUG] Received EC2 Placement Group: %s", pg)

	d.Set("name", pg.GroupName)
	d.Set("strategy", pg.Strategy)
	d.Set("placement_group_id", pg.GroupId)
	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(pg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsPlacementGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		pgId := d.Get("placement_group_id").(string)
		if err := keyvaluetags.Ec2UpdateTags(conn, pgId, o, n); err != nil {
			return fmt.Errorf("error updating Placement Group (%s) tags: %s", pgId, err)
		}
	}

	return resourceAwsPlacementGroupRead(d, meta)
}

func resourceAwsPlacementGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Deleting EC2 Placement Group %q", d.Id())
	_, err := conn.DeletePlacementGroup(&ec2.DeletePlacementGroupInput{
		GroupName: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	wait := resource.StateChangeConf{
		Pending:    []string{ec2.PlacementGroupStateAvailable, ec2.PlacementGroupStateDeleting},
		Target:     []string{ec2.PlacementGroupStateDeleted},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
				GroupNames: []*string{aws.String(d.Id())},
			})

			if err != nil {
				if isAWSErr(err, "InvalidPlacementGroup.Unknown", "") {
					return out, ec2.PlacementGroupStateDeleted, nil
				}
				return out, "", err
			}

			if len(out.PlacementGroups) == 0 {
				return out, ec2.PlacementGroupStateDeleted, nil
			}

			pg := out.PlacementGroups[0]
			if *pg.State == ec2.PlacementGroupStateAvailable {
				log.Printf("[DEBUG] Accepted status when deleting EC2 Placement group: %q %v", d.Id(), *pg.State)
			}

			return out, *pg.State, nil
		},
	}

	_, err = wait.WaitForState()
	return err
}
