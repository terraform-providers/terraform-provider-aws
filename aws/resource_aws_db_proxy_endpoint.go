package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/waiter"
)

func resourceAwsDbProxyEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbProxyEndpointCreate,
		Read:   resourceAwsDbProxyEndpointRead,
		Delete: resourceAwsDbProxyEndpointDelete,
		Update: resourceAwsDbProxyEndpointUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: SetTagsDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"db_proxy_endpoint_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"target_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      rds.DBProxyEndpointTargetRoleReadWrite,
				ValidateFunc: validation.StringInSlice(rds.DBProxyEndpointTargetRole_Values(), false),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsDbProxyEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	dbProxyName := d.Get("db_proxy_name").(string)
	dbProxyEndpointName := d.Get("db_proxy_endpoint_name").(string)

	params := rds.CreateDBProxyEndpointInput{
		DBProxyName:         aws.String(dbProxyName),
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
		TargetRole:          aws.String(d.Get("target_role").(string)),
		VpcSubnetIds:        expandStringSet(d.Get("vpc_subnet_ids").(*schema.Set)),
		Tags:                tags.IgnoreAws().RdsTags(),
	}

	if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
		params.VpcSecurityGroupIds = expandStringSet(v)
	}

	_, err := conn.CreateDBProxyEndpoint(&params)

	if err != nil {
		return fmt.Errorf("error Creating RDS DB Proxy Endpoint (%s/%s): %w", dbProxyName, dbProxyEndpointName, err)
	}

	d.SetId(strings.Join([]string{dbProxyName, dbProxyEndpointName}, "/"))

	if _, err := waiter.DBProxyEndpointAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for RDS DB Proxy Endpoint (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsDbProxyEndpointRead(d, meta)
}

func resourceAwsDbProxyEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	dbProxyEndpoint, err := finder.DBProxyEndpoint(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Proxy Endpoint (%s): %w", d.Id(), err)
	}

	if dbProxyEndpoint == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading RDS DB Proxy Endpoint (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] RDS DB Proxy Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	endpointArn := aws.StringValue(dbProxyEndpoint.DBProxyEndpointArn)
	d.Set("arn", endpointArn)
	d.Set("db_proxy_name", dbProxyEndpoint.DBProxyName)
	d.Set("endpoint", dbProxyEndpoint.Endpoint)
	d.Set("db_proxy_endpoint_name", dbProxyEndpoint.DBProxyEndpointName)
	d.Set("is_default", dbProxyEndpoint.IsDefault)
	d.Set("target_role", dbProxyEndpoint.TargetRole)
	d.Set("vpc_id", dbProxyEndpoint.VpcId)
	d.Set("target_role", dbProxyEndpoint.TargetRole)
	d.Set("vpc_subnet_ids", flattenStringSet(dbProxyEndpoint.VpcSubnetIds))
	d.Set("vpc_security_group_ids", flattenStringSet(dbProxyEndpoint.VpcSecurityGroupIds))

	tags, err := keyvaluetags.RdsListTags(conn, endpointArn)

	if err != nil {
		return fmt.Errorf("Error listing tags for RDS DB Proxy Endpoint (%s): %w", endpointArn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsDbProxyEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	if d.HasChange("vpc_security_group_ids") {
		params := rds.ModifyDBProxyEndpointInput{
			DBProxyEndpointName: aws.String(d.Get("db_proxy_endpoint_name").(string)),
			VpcSecurityGroupIds: expandStringSet(d.Get("vpc_security_group_ids").(*schema.Set)),
		}

		_, err := conn.ModifyDBProxyEndpoint(&params)
		if err != nil {
			return fmt.Errorf("Error updating DB Proxy Endpoint: %w", err)
		}

		if _, err := waiter.DBProxyEndpointAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for RDS DB Proxy Endpoint (%s) to become modified: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RdsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating RDS DB Proxy Endpoint (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsDbProxyEndpointRead(d, meta)
}

func resourceAwsDbProxyEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := rds.DeleteDBProxyEndpointInput{
		DBProxyEndpointName: aws.String(d.Get("db_proxy_endpoint_name").(string)),
	}

	log.Printf("[DEBUG] Delete DB Proxy Endpoint: %#v", params)
	_, err := conn.DeleteDBProxyEndpoint(&params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) || tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
			return nil
		}
		return fmt.Errorf("Error Deleting DB Proxy Endpoint: %w", err)
	}

	if _, err := waiter.DBProxyEndpointDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) || tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
			return nil
		}
		return fmt.Errorf("error waiting for RDS DB Proxy Endpoint (%s) to become deleted: %w", d.Id(), err)
	}

	return nil
}
