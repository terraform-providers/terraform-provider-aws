package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceAwsElasticacheUserGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsElasticacheUserGroupCreate,
		Read:   resourceAwsElasticacheUserGroupRead,
		Update: resourceAwsElasticacheUserGroupUpdate,
		Delete: resourceAwsElasticacheUserGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 40),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z-]+$`), "must contain only alphanumeric characters and hyphens"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with a letter"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "redis",
				ValidateFunc: validateAwsElastiCacheUserGroupEngine,
			},
			"user_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
		SchemaVersion: 1,

		// SchemaVersion: 1 did not include any state changes via MigrateState.
		// Perform a no-operation state upgrade for Terraform 0.12 compatibility.
		// Future state migrations should be performed with StateUpgraders.
		MigrateState: func(v int, inst *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
			return inst, nil
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
		},
	}
}

func resourceAwsElasticacheUserGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	userGroupId := d.Get("user_group_id").(string)

	// Get ElastiCache User Properties
	params := &elasticache.CreateUserGroupInput{
		Engine:      aws.String(d.Get("engine").(string)),
		UserGroupId: aws.String(d.Get("user_group_id").(string)),
	}

	if userIds := d.Get("user_ids").(*schema.Set); userIds.Len() > 0 {
		params.UserIds = expandStringSet(userIds)
	}

	_, err := conn.CreateUserGroup(params)
	if err != nil {
		return fmt.Errorf("[ERROR] Error creating ElastiCache User Group: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"active"},
		Refresh:    resourceAwsElasticacheUserGroupStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      10 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for Elasticache User Group (%s) to become available", d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[ERROR] Elasticache User Group (%s) during creation: %w", d.Id(), err)
	}

	d.SetId(userGroupId)

	return resourceAwsElasticacheUserGroupRead(d, meta)
}

func resourceAwsElasticacheUserGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	params := &elasticache.DescribeUserGroupsInput{
		UserGroupId: aws.String(d.Get("user_group_id").(string)),
	}

	response, err := conn.DescribeUserGroups(params)

	if isAWSErr(err, elasticache.ErrCodeUserGroupNotFoundFault, "") {
		log.Printf("[WARN] ElastiCache User Group (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("[ERROR] ElastiCache User Group Reading: %s", err)
	}

	if response == nil || len(response.UserGroups) == 0 || response.UserGroups[0] == nil {
		log.Printf("[WARN] ElastiCache User Groups (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	userGroup := response.UserGroups[0]

	d.Set("engine", userGroup.Engine)
	d.Set("user_ids", userGroup.UserIds)
	d.Set("user_group_id", userGroup.UserGroupId)

	return nil
}

func resourceAwsElasticacheUserGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	if d.HasChanges("user_ids") {
		o, n := d.GetChange("user_ids")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		addUsers := ns.Difference(os).List()
		removeUsers := os.Difference(ns).List()

		if len(addUsers) > 0 {
			params := &elasticache.ModifyUserGroupInput{
				UserGroupId:  aws.String(d.Get("user_group_id").(string)),
				UserIdsToAdd: expandStringList(addUsers),
			}

			_, err := conn.ModifyUserGroup(params)
			if err != nil {
				return err
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{"modifying"},
				Target:     []string{"active"},
				Refresh:    resourceAwsElasticacheUserGroupStateRefreshFunc(conn, d.Id()),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				MinTimeout: 10 * time.Second,
				Delay:      10 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for Elasticache User Group (%s) to become available", d.Id())

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("[ERROR] Elasticache User Group (%s) during modification: %w", d.Id(), err)
			}
		} else if len(removeUsers) > 0 {
			params := &elasticache.ModifyUserGroupInput{
				UserGroupId:     aws.String(d.Get("user_group_id").(string)),
				UserIdsToRemove: expandStringList(removeUsers),
			}

			_, err := conn.ModifyUserGroup(params)
			if err != nil {
				return err
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{"modifying"},
				Target:     []string{"active"},
				Refresh:    resourceAwsElasticacheUserGroupStateRefreshFunc(conn, d.Id()),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				MinTimeout: 10 * time.Second,
				Delay:      10 * time.Second,
			}

			log.Printf("[DEBUG] Waiting for Elasticache User Group (%s) to become available", d.Id())

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("[ERROR] Elasticache User Group (%s) during modification: %w", d.Id(), err)
			}
		}
	}
	return resourceAwsElasticacheUserGroupRead(d, meta)
}

func resourceAwsElasticacheUserGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	params := &elasticache.DeleteUserGroupInput{
		UserGroupId: aws.String(d.Get("user_group_id").(string)),
	}

	log.Printf("[DEBUG] ElastiCache User Group Delete: %#v", params)

	_, err := conn.DeleteUserGroup(params)

	if isAWSErr(err, elasticache.ErrCodeUserGroupNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("[ERROR] Deleting ElastiCache User Group (%s): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"deleting"},
		Target:     []string{},
		Refresh:    resourceAwsElasticacheUserGroupStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 30 * time.Second,
		Delay:      60 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[ERROR] ElastiCache User Group (%s) during Deletion: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsElasticacheUserGroupStateRefreshFunc(conn *elasticache.ElastiCache, userGroupId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		params := &elasticache.DescribeUserGroupsInput{
			UserGroupId: aws.String(userGroupId),
		}

		response, err := conn.DescribeUserGroups(params)

		if isAWSErr(err, elasticache.ErrCodeUserGroupNotFoundFault, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if response == nil || len(response.UserGroups) == 0 || response.UserGroups[0] == nil {
			return nil, "", nil
		}

		return response, aws.StringValue(response.UserGroups[0].Status), nil
	}
}

func validateAwsElastiCacheUserGroupEngine(v interface{}, k string) (ws []string, errors []error) {
	if strings.ToLower(v.(string)) != "redis" {
		errors = append(errors, fmt.Errorf("The only acceptable Engine type when using User Groups is Redis"))
	}
	return
}
