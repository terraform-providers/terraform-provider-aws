package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudhsmv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudhsmv2/waiter"
)

func resourceAwsCloudHsmV2Hsm() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudHsmV2HsmCreate,
		Read:   resourceAwsCloudHsmV2HsmRead,
		Delete: resourceAwsCloudHsmV2HsmDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"hsm_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hsm_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hsm_eni_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudHsmV2HsmCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn

	input := &cloudhsmv2.CreateHsmInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		input.AvailabilityZone = aws.String(v.(string))
	} else {
		cluster, err := finder.Cluster(conn, d.Get("cluster_id").(string))

		if err != nil {
			return fmt.Errorf("error reading CloudHSMv2 Cluster (%s): %w", d.Id(), err)
		}

		if cluster == nil {
			return fmt.Errorf("error reading CloudHSMv2 Cluster (%s): not found for subnet mappings", d.Id())
		}

		subnetId := d.Get("subnet_id").(string)
		for az, sn := range cluster.SubnetMapping {
			if aws.StringValue(sn) == subnetId {
				input.AvailabilityZone = aws.String(az)
			}
		}
	}

	if v, ok := d.GetOk("ip_address"); ok {
		input.IpAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] CloudHSMv2 HSM create %s", input)

	output, err := conn.CreateHsm(input)

	if err != nil {
		return fmt.Errorf("error creating CloudHSMv2 HSM: %w", err)
	}

	d.SetId(aws.StringValue(output.Hsm.HsmId))

	if _, err := waiter.HsmActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for CloudHSMv2 HSM (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsCloudHsmV2HsmRead(d, meta)
}

func resourceAwsCloudHsmV2HsmRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn

	hsm, err := finder.Hsm(conn, d.Id(), d.Get("hsm_eni_id").(string))

	if err != nil {
		return fmt.Errorf("error reading CloudHSMv2 HSM (%s): %w", d.Id(), err)
	}

	if hsm == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading CloudHSMv2 HSM (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] CloudHSMv2 HSM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// When matched by ENI ID, the ID should updated.
	if aws.StringValue(hsm.HsmId) != d.Id() {
		d.SetId(aws.StringValue(hsm.HsmId))
	}

	log.Printf("[INFO] Reading CloudHSMv2 HSM Information: %s", d.Id())

	d.Set("cluster_id", hsm.ClusterId)
	d.Set("subnet_id", hsm.SubnetId)
	d.Set("availability_zone", hsm.AvailabilityZone)
	d.Set("ip_address", hsm.EniIp)
	d.Set("hsm_id", hsm.HsmId)
	d.Set("hsm_state", hsm.State)
	d.Set("hsm_eni_id", hsm.EniId)

	return nil
}

func resourceAwsCloudHsmV2HsmDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn
	clusterId := d.Get("cluster_id").(string)

	log.Printf("[DEBUG] CloudHSMv2 HSM delete %s %s", clusterId, d.Id())
	input := &cloudhsmv2.DeleteHsmInput{
		ClusterId: aws.String(clusterId),
		HsmId:     aws.String(d.Id()),
	}

	_, err := conn.DeleteHsm(input)

	if tfawserr.ErrCodeEquals(err, cloudhsmv2.ErrCodeCloudHsmResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudHSMv2 HSM (%s): %w", d.Id(), err)
	}

	if _, err := waiter.HsmDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for CloudHSMv2 HSM (%s) deletion: %w", d.Id(), err)
	}

	return nil
}
