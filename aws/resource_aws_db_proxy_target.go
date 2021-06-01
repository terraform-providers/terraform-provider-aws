package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
)

func resourceAwsDbProxyTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbProxyTargetCreate,
		Read:   resourceAwsDbProxyTargetRead,
		Delete: resourceAwsDbProxyTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"target_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
				ValidateFunc: validateRdsIdentifier,
			},
			"db_cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
				ValidateFunc: validateRdsIdentifier,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"rds_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tracked_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDbProxyTargetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbProxyName := d.Get("db_proxy_name").(string)
	targetGroupName := d.Get("target_group_name").(string)

	params := rds.RegisterDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		params.DBInstanceIdentifiers = []*string{aws.String(v.(string))}
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		params.DBClusterIdentifiers = []*string{aws.String(v.(string))}
	}

	resp, err := conn.RegisterDBProxyTargets(&params)

	if err != nil {
		return fmt.Errorf("error registering RDS DB Proxy (%s/%s) Target: %w", dbProxyName, targetGroupName, err)
	}

	dbProxyTarget := resp.DBProxyTargets[0]

	d.SetId(strings.Join([]string{dbProxyName, targetGroupName, aws.StringValue(dbProxyTarget.Type), aws.StringValue(dbProxyTarget.RdsResourceId)}, "/"))

	return resourceAwsDbProxyTargetRead(d, meta)
}

func resourceAwsDbProxyTargetParseID(id string) (string, string, string, string, error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%s), expected db_proxy_name/target_group_name/type/id", id)
	}
	return idParts[0], idParts[1], idParts[2], idParts[3], nil
}

func resourceAwsDbProxyTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbProxyName, targetGroupName, targetType, rdsResourceId, err := resourceAwsDbProxyTargetParseID(d.Id())
	if err != nil {
		return err
	}

	dbProxyTarget, err := finder.DBProxyTarget(conn, dbProxyName, targetGroupName, targetType, rdsResourceId)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetGroupNotFoundFault) {
		log.Printf("[WARN] RDS DB Proxy Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Proxy Target (%s): %w", d.Id(), err)
	}

	if dbProxyTarget == nil {
		log.Printf("[WARN] RDS DB Proxy Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("db_proxy_name", dbProxyName)
	d.Set("endpoint", dbProxyTarget.Endpoint)
	d.Set("port", dbProxyTarget.Port)
	d.Set("rds_resource_id", dbProxyTarget.RdsResourceId)
	d.Set("target_arn", dbProxyTarget.TargetArn)
	d.Set("target_group_name", targetGroupName)
	d.Set("tracked_cluster_id", dbProxyTarget.TrackedClusterId)
	d.Set("type", dbProxyTarget.Type)

	if aws.StringValue(dbProxyTarget.Type) == rds.TargetTypeRdsInstance {
		d.Set("db_instance_identifier", dbProxyTarget.RdsResourceId)
	} else {
		d.Set("db_cluster_identifier", dbProxyTarget.RdsResourceId)
	}

	return nil
}

func resourceAwsDbProxyTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := rds.DeregisterDBProxyTargetsInput{
		DBProxyName:     aws.String(d.Get("db_proxy_name").(string)),
		TargetGroupName: aws.String(d.Get("target_group_name").(string)),
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		params.DBInstanceIdentifiers = []*string{aws.String(v.(string))}
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		params.DBClusterIdentifiers = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Deregister DB Proxy target: %#v", params)
	_, err := conn.DeregisterDBProxyTargets(&params)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetGroupNotFoundFault) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyTargetNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deregistering DB Proxy target: %s", err)
	}

	return nil
}
