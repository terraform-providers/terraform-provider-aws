package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Constants not currently provided by the AWS Go SDK
const (
	rdsDbInstanceRoleStatusActive  = "ACTIVE"
	rdsDbInstanceRoleStatusDeleted = "DELETED"
	rdsDbInstanceRoleStatusPending = "PENDING"
)

func resourceAwsDbInstanceRoleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbInstanceRoleAssociationCreate,
		Read:   resourceAwsDbInstanceRoleAssociationRead,
		Delete: resourceAwsDbInstanceRoleAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsDbInstanceRoleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbInstanceIdentifier := d.Get("db_instance_identifier").(string)
	roleArn := d.Get("role_arn").(string)

	input := &rds.AddRoleToDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleArn),
	}

	log.Printf("[DEBUG] RDS DB Instance (%s) IAM Role associating: %s", dbInstanceIdentifier, roleArn)
	_, err := conn.AddRoleToDBInstance(input)

	if err != nil {
		return fmt.Errorf("error associating RDS DB Instance (%s) IAM Role (%s): %s", dbInstanceIdentifier, roleArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", dbInstanceIdentifier, roleArn))

	if err := waitForRdsDbInstanceRoleAssociation(conn, dbInstanceIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Instance (%s) IAM Role (%s) association: %s", dbInstanceIdentifier, roleArn, err)
	}

	return resourceAwsDbInstanceRoleAssociationRead(d, meta)
}

func resourceAwsDbInstanceRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbInstanceIdentifier, roleArn, err := resourceAwsDbInstanceRoleAssociationDecodeId(d.Id())

	if err != nil {
		return fmt.Errorf("error reading resource ID: %s", err)
	}

	dbInstanceRole, err := rdsDescribeDbInstanceRole(conn, dbInstanceIdentifier, roleArn)

	if isAWSErr(err, rds.ErrCodeDBInstanceNotFoundFault, "") {
		log.Printf("[WARN] RDS DB Instance (%s) not found, removing from state", dbInstanceIdentifier)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS DB Instance (%s) IAM Role (%s) association: %s", dbInstanceIdentifier, roleArn, err)
	}

	if dbInstanceRole == nil {
		log.Printf("[WARN] RDS DB Instance (%s) IAM Role (%s) association not found, removing from state", dbInstanceIdentifier, roleArn)
		d.SetId("")
		return nil
	}

	d.Set("db_instance_identifier", dbInstanceIdentifier)
	d.Set("feature_name", dbInstanceRole.FeatureName)
	d.Set("role_arn", dbInstanceRole.RoleArn)

	return nil
}

func resourceAwsDbInstanceRoleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbInstanceIdentifier, roleArn, err := resourceAwsDbInstanceRoleAssociationDecodeId(d.Id())

	if err != nil {
		return fmt.Errorf("error reading resource ID: %s", err)
	}

	input := &rds.RemoveRoleFromDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		FeatureName:          aws.String(d.Get("feature_name").(string)),
		RoleArn:              aws.String(roleArn),
	}

	log.Printf("[DEBUG] RDS DB Instance (%s) IAM Role disassociating: %s", dbInstanceIdentifier, roleArn)
	_, err = conn.RemoveRoleFromDBInstance(input)

	if isAWSErr(err, rds.ErrCodeDBInstanceNotFoundFault, "") {
		return nil
	}

	if isAWSErr(err, rds.ErrCodeDBInstanceRoleNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RDS DB Instance (%s) IAM Role (%s): %s", dbInstanceIdentifier, roleArn, err)
	}

	if err := waitForRdsDbInstanceRoleDisassociation(conn, dbInstanceIdentifier, roleArn); err != nil {
		return fmt.Errorf("error waiting for RDS DB Instance (%s) IAM Role (%s) disassociation: %s", dbInstanceIdentifier, roleArn, err)
	}

	return nil
}

func resourceAwsDbInstanceRoleAssociationDecodeId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected DB-INSTANCE-ID,ROLE-ARN", id)
	}

	return parts[0], parts[1], nil
}

func rdsDescribeDbInstanceRole(conn *rds.RDS, dbInstanceIdentifier, roleArn string) (*rds.DBInstanceRole, error) {
	input := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	}

	log.Printf("[DEBUG] Describing RDS DB Instance: %s", input)
	output, err := conn.DescribeDBInstances(input)

	if err != nil {
		return nil, err
	}

	var dbInstance *rds.DBInstance

	for _, outputDbInstance := range output.DBInstances {
		if aws.StringValue(outputDbInstance.DBInstanceIdentifier) == dbInstanceIdentifier {
			dbInstance = outputDbInstance
			break
		}
	}

	if dbInstance == nil {
		return nil, nil
	}

	var dbInstanceRole *rds.DBInstanceRole

	for _, associatedRole := range dbInstance.AssociatedRoles {
		if aws.StringValue(associatedRole.RoleArn) == roleArn {
			dbInstanceRole = associatedRole
			break
		}
	}

	return dbInstanceRole, nil
}

func waitForRdsDbInstanceRoleAssociation(conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rdsDbInstanceRoleStatusPending},
		Target:  []string{rdsDbInstanceRoleStatusActive},
		Refresh: func() (interface{}, string, error) {
			dbInstanceRole, err := rdsDescribeDbInstanceRole(conn, dbInstanceIdentifier, roleArn)

			if err != nil {
				return nil, "", err
			}

			if dbInstanceRole == nil {
				return nil, rdsDbInstanceRoleStatusPending, nil
			}

			return dbInstanceRole, aws.StringValue(dbInstanceRole.Status), nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role association: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}

func waitForRdsDbInstanceRoleDisassociation(conn *rds.RDS, dbInstanceIdentifier, roleArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			rdsDbInstanceRoleStatusActive,
			rdsDbInstanceRoleStatusPending,
		},
		Target: []string{rdsDbInstanceRoleStatusDeleted},
		Refresh: func() (interface{}, string, error) {
			dbInstanceRole, err := rdsDescribeDbInstanceRole(conn, dbInstanceIdentifier, roleArn)

			if isAWSErr(err, rds.ErrCodeDBInstanceNotFoundFault, "") {
				return &rds.DBInstanceRole{}, rdsDbInstanceRoleStatusDeleted, nil
			}

			if err != nil {
				return nil, "", err
			}

			if dbInstanceRole != nil {
				return dbInstanceRole, aws.StringValue(dbInstanceRole.Status), nil
			}

			return &rds.DBInstanceRole{}, rdsDbInstanceRoleStatusDeleted, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for RDS DB Instance (%s) IAM Role disassociation: %s", dbInstanceIdentifier, roleArn)
	_, err := stateConf.WaitForState()

	return err
}
