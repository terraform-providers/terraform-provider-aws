package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsEfsMountTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEfsMountTargetCreate,
		Read:   resourceAwsEfsMountTargetRead,
		Update: resourceAwsEfsMountTargetUpdate,
		Delete: resourceAwsEfsMountTargetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ip_address": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceAwsEfsMountTargetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	fsId := d.Get("file_system_id").(string)
	subnetId := d.Get("subnet_id").(string)

	// CreateMountTarget would return the same Mount Target ID
	// to parallel requests if they both include the same AZ
	// and we would end up managing the same MT as 2 resources.
	// So we make it fail by calling 1 request per AZ at a time.
	az, err := getAzFromSubnetId(subnetId, meta.(*AWSClient).ec2conn)
	if err != nil {
		return fmt.Errorf("Failed getting Availability Zone from subnet ID (%s): %s", subnetId, err)
	}
	mtKey := "efs-mt-" + fsId + "-" + az
	awsMutexKV.Lock(mtKey)
	defer awsMutexKV.Unlock(mtKey)

	input := efs.CreateMountTargetInput{
		FileSystemId: aws.String(fsId),
		SubnetId:     aws.String(subnetId),
	}

	if v, ok := d.GetOk("ip_address"); ok {
		input.IpAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("security_groups"); ok {
		input.SecurityGroups = expandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating EFS mount target: %#v", input)

	mt, err := conn.CreateMountTarget(&input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(mt.MountTargetId))
	log.Printf("[INFO] EFS mount target ID: %s", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateCreating},
		Target:  []string{efs.LifeCycleStateAvailable},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				MountTargetId: aws.String(d.Id()),
			})
			if err != nil {
				return nil, "error", err
			}

			if hasEmptyMountTargets(resp) {
				return nil, "error", fmt.Errorf("EFS mount target %q could not be found.", d.Id())
			}

			mt := resp.MountTargets[0]

			log.Printf("[DEBUG] Current status of %q: %q", aws.StringValue(mt.MountTargetId), aws.StringValue(mt.LifeCycleState))
			return mt, aws.StringValue(mt.LifeCycleState), nil
		},
		Timeout:    10 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for EFS mount target (%s) to create: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS mount target created: %s", aws.StringValue(mt.MountTargetId))

	return resourceAwsEfsMountTargetRead(d, meta)
}

func resourceAwsEfsMountTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	if d.HasChange("security_groups") {
		input := efs.ModifyMountTargetSecurityGroupsInput{
			MountTargetId:  aws.String(d.Id()),
			SecurityGroups: expandStringSet(d.Get("security_groups").(*schema.Set)),
		}
		_, err := conn.ModifyMountTargetSecurityGroups(&input)
		if err != nil {
			return err
		}
	}

	return resourceAwsEfsMountTargetRead(d, meta)
}

func resourceAwsEfsMountTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn
	resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
		MountTargetId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, efs.ErrCodeMountTargetNotFound, "") {
			// The EFS mount target could not be found,
			// which would indicate that it might be
			// already deleted.
			log.Printf("[WARN] EFS mount target %q could not be found.", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading EFS mount target %s: %s", d.Id(), err)
	}

	if hasEmptyMountTargets(resp) {
		return fmt.Errorf("EFS mount target %q could not be found.", d.Id())
	}

	mt := resp.MountTargets[0]

	log.Printf("[DEBUG] Found EFS mount target: %#v", mt)

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

func getAzFromSubnetId(subnetId string, conn *ec2.EC2) (string, error) {
	input := ec2.DescribeSubnetsInput{
		SubnetIds: []*string{aws.String(subnetId)},
	}
	out, err := conn.DescribeSubnets(&input)
	if err != nil {
		return "", err
	}

	if l := len(out.Subnets); l != 1 {
		return "", fmt.Errorf("Expected exactly 1 subnet returned for %q, got: %d", subnetId, l)
	}

	return aws.StringValue(out.Subnets[0].AvailabilityZone), nil
}

func resourceAwsEfsMountTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	log.Printf("[DEBUG] Deleting EFS mount target %q", d.Id())
	_, err := conn.DeleteMountTarget(&efs.DeleteMountTargetInput{
		MountTargetId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	err = waitForDeleteEfsMountTarget(conn, d.Id(), 10*time.Minute)
	if err != nil {
		return fmt.Errorf("Error waiting for EFS mount target (%q) to delete: %s", d.Id(), err.Error())
	}

	log.Printf("[DEBUG] EFS mount target %q deleted.", d.Id())

	return nil
}

func waitForDeleteEfsMountTarget(conn *efs.EFS, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting, efs.LifeCycleStateDeleted},
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				MountTargetId: aws.String(id),
			})
			if err != nil {
				if isAWSErr(err, efs.ErrCodeMountTargetNotFound, "") {
					return nil, "", nil
				}

				return nil, "error", err
			}

			if hasEmptyMountTargets(resp) {
				return nil, "", nil
			}

			mt := resp.MountTargets[0]

			log.Printf("[DEBUG] Current status of %q: %q", aws.StringValue(mt.MountTargetId), aws.StringValue(mt.LifeCycleState))
			return mt, aws.StringValue(mt.LifeCycleState), nil
		},
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func hasEmptyMountTargets(mto *efs.DescribeMountTargetsOutput) bool {
	if mto != nil && len(mto.MountTargets) > 0 {
		return false
	}
	return true
}
