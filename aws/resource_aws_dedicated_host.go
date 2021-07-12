package aws

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDedicatedHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDedicatedHostCreate,
		Read:   resourceAwsDedicatedHostRead,
		Update: resourceAwsDedicatedHostUpdate,
		Delete: resourceAwsDedicatedHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tags": tagsSchema(),
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host_recovery": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.HostRecoveryOn,
					ec2.HostRecoveryOff,
				}, false),
			},
			"auto_placement": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AutoPlacementOn,
					ec2.AutoPlacementOff,
				}, false),
			},
		},
	}
}

type awsHostsOpts struct {
	AutoPlacement    *string
	AvailabilityZone *string
	InstanceType     *string
	HostRecovery     *string
}

func buildAwsHostsOpts(d *schema.ResourceData) *awsHostsOpts {

	instanceType := d.Get("instance_type").(string)
	opts := &awsHostsOpts{
		AutoPlacement:    aws.String(d.Get("auto_placement").(string)),
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		InstanceType:     aws.String(instanceType),
		HostRecovery:     aws.String(d.Get("host_recovery").(string)),
	}
	return opts
}

// resourceAwsDedicatedHostCreate allocates a Dedicated Host to your account.
// https://docs.aws.amazon.com/en_pv/AWSEC2/latest/APIReference/API_AllocateHosts.html
func resourceAwsDedicatedHostCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	hostOpts := buildAwsHostsOpts(d)

	tagsSpec := ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeDedicatedHost)

	// Build the creation struct
	runOpts := &ec2.AllocateHostsInput{
		AvailabilityZone: hostOpts.AvailabilityZone,
		Quantity:         aws.Int64(int64(1)),
		InstanceType:     hostOpts.InstanceType,
		HostRecovery:     hostOpts.HostRecovery,
		AutoPlacement:    hostOpts.AutoPlacement,
	}

	if len(tagsSpec) > 0 {
		runOpts.TagSpecifications = tagsSpec
	}

	var runResp *ec2.AllocateHostsOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		runResp, err = conn.AllocateHosts(runOpts)
		return resource.RetryableError(err)
	})
	if isResourceTimeoutError(err) {
		runResp, err = conn.AllocateHosts(runOpts)
	}
	if err != nil {
		return fmt.Errorf("Error launching host : %s", err)
	}
	if runResp == nil || len(runResp.HostIds) == 0 {
		return errors.New("Error launching source host: no hosts returned in response")
	}

	log.Printf("[INFO] Host ID: %s", *runResp.HostIds[0])
	d.SetId(*runResp.HostIds[0])

	// Update if we need to
	return resourceAwsDedicatedHostRead(d, meta)
}

func resourceAwsDedicatedHostRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeHosts(&ec2.DescribeHostsInput{
		HostIds: []*string{aws.String(d.Id())},
	})
	if err != nil {
		// If the host was not found, return nil so that we can show
		// that host is gone.
		if isAWSErr(err, "InvalidHostID.NotFound", "") {
			d.SetId("")
			return nil
		}

		// Some other error, report it
		return err
	}
	if len(resp.Hosts) == 0 {
		d.SetId("")
		return nil
	}
	host := resp.Hosts[0]
	d.Set("auto_placement", host.AutoPlacement)
	d.Set("availibility_zone", host.AvailabilityZone)
	d.Set("host_recovery", host.HostRecovery)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(host.Tags).IgnoreConfig(ignoreTagsConfig).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	// If nothing was found, then return no state

	return nil
}

// resourceAwsDedicatedHostUpdate modifies AWS Host AutoPlacement and HostRecovery settings.
// When auto-placement is enabled, any instances that you launch with a tenancy of host but without a specific host ID are placed onto any available
// Dedicated Host in your account that has auto-placement enabled.
// https://docs.aws.amazon.com/en_pv/AWSEC2/latest/APIReference/API_ModifyHosts.html
func resourceAwsDedicatedHostUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	requestUpdate := false
	req := &ec2.ModifyHostsInput{
		HostIds: []*string{aws.String(d.Id())},
	}
	if d.HasChange("auto_placement") {
		req.AutoPlacement = aws.String(d.Get("auto_placement").(string))
		requestUpdate = true
	}
	// Indicates whether to enable or disable host recovery for the Dedicated Host.
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/dedicated-hosts-recovery.html
	if d.HasChange("host_recovery") {
		req.HostRecovery = aws.String(d.Get("host_recovery").(string))
		requestUpdate = true
	}
	if requestUpdate {
		err := resource.Retry(30*time.Second, func() *resource.RetryError {
			_, err := conn.ModifyHosts(req)
			if err != nil {
				return resource.RetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Error modifying host %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsDedicatedHostRead(d, meta)
}

func resourceAwsDedicatedHostDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	err := awsReleaseHosts(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error terminating EC2 Host (%s): %s", d.Id(), err)
	}

	return nil
}

func awsReleaseHosts(conn *ec2.EC2, id string) error {
	log.Printf("[INFO] Terminating host: %s", id)
	req := &ec2.ReleaseHostsInput{
		HostIds: []*string{aws.String(id)},
	}
	if _, err := conn.ReleaseHosts(req); err != nil {
		if isAWSErr(err, "InvalidHostID.NotFound", "") {
			return nil
		}
		return err
	}

	return nil
}
