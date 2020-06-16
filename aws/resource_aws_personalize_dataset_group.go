package aws

import (
	//"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/personalize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"time"
)

func resourceAwsPersonalizeDatasetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPersonalizeDatasetGroupCreate,
		Read:   resourceAwsPersonalizeDatasetGroupRead,
		Delete: resourceAwsPersonalizeDatasetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"kms": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},

						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
		},
	}
}

func resourceAwsPersonalizeDatasetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).personalizeconn

	createOpts := &personalize.CreateDatasetGroupInput{
		Name: aws.String(d.Get("name").(string)),
	}

	kms := d.Get("kms").([]interface{})

	if len(kms) > 0 {
		var k map[string]interface{}
		if kms[0] != nil {
			k = kms[0].(map[string]interface{})
		} else {
			k = make(map[string]interface{})
		}

		if v, ok := k["key_arn"]; ok {
			createOpts.KmsKeyArn = aws.String(v.(string))
		}

		if v, ok := k["role_arn"]; ok {
			createOpts.RoleArn = aws.String(v.(string))
		}
	}

	log.Printf("[DEBUG] Personalize dataset group create options: %#v", *createOpts)

	ds, err := conn.CreateDatasetGroup(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Personalize dataset group: %s", err)
	}

	d.SetId(*ds.DatasetGroupArn)
	log.Printf("[INFO] Personalize dataset group ARN: %s", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"CREATE PENDING", "CREATE IN_PROGRESS"},
		Target:     []string{"ACTIVE"},
		Refresh:    resourceAwsPersonalizeDatasetGroupCreateRefreshFunc(d.Id(), conn),
		Timeout:    10 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Personalize dataset group (%q) to create: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Personalize dataset group %q created.", d.Id())

	return resourceAwsPersonalizeDatasetGroupRead(d, meta)
}

func resourceAwsPersonalizeDatasetGroupCreateRefreshFunc(id string, conn *personalize.Personalize) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDatasetGroup(&personalize.DescribeDatasetGroupInput{
			DatasetGroupArn: aws.String(id),
		})
		if err != nil {
			return nil, "error", err
		}

		if resp == nil || resp.DatasetGroup == nil {
			return nil, "not-found", fmt.Errorf("Personalize dataset group %q could not be found.", id)
		}

		dg := resp.DatasetGroup
		state := aws.StringValue(dg.Status)
		log.Printf("[DEBUG] current status of %q: %q", id, state)
		return dg, state, nil
	}
}

func resourceAwsPersonalizeDatasetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).personalizeconn

	if _, err := arn.Parse(d.Id()); err != nil {
		d.SetId(arn.ARN{
			AccountID: meta.(*AWSClient).accountid,
			Partition: meta.(*AWSClient).partition,
			Region:    meta.(*AWSClient).region,
			Resource:  fmt.Sprintf("dataset-group/%s", d.Id()),
			Service:   "personalize",
		}.String())
	}

	resp, err := conn.DescribeDatasetGroup(&personalize.DescribeDatasetGroupInput{
		DatasetGroupArn: aws.String(d.Id()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Personalize dataset group (%s) could not be found.", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.DatasetGroup == nil {
		return fmt.Errorf("Personalize dataset group %q could not be found.", d.Id())
	}

	dg := resp.DatasetGroup

	// In the import case, we won't have this
	if _, ok := d.GetOk("arn"); !ok {
		d.Set("arn", d.Id())
	}

	d.Set("name", dg.Name)
	d.Set("kms_key_arn", dg.KmsKeyArn)
	d.Set("role_arn", dg.RoleArn)

	return nil
}

func resourceAwsPersonalizeDatasetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).personalizeconn

	log.Printf("[DEBUG] Deleting Personalize dataset group: %s", d.Id())
	_, err := conn.DeleteDatasetGroup(&personalize.DeleteDatasetGroupInput{
		DatasetGroupArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting dataset group: %s with err %s", d.Id(), err.Error())
	}

	log.Printf("[DEBUG] Personalize dataset group %q deleted.", d.Id())

	return nil
}
