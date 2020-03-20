package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func resourceAwsCloudHsmV2Hsm() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudHsmV2HsmCreate,
		Read:   resourceAwsCloudHsmV2HsmRead,
		Delete: resourceAwsCloudHsmV2HsmDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsCloudHsmV2HsmImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
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

func resourceAwsCloudHsmV2HsmImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("hsm_id", d.Id())
	return []*schema.ResourceData{d}, nil
}

func describeHsm(conn *cloudhsmv2.CloudHSMV2, hsmId string) (*cloudhsmv2.Hsm, error) {
	out, err := conn.DescribeClusters(&cloudhsmv2.DescribeClustersInput{})
	if err != nil {
		log.Printf("[WARN] Error on descibing CloudHSM v2 Cluster: %s", err)
		return nil, err
	}

	var hsm *cloudhsmv2.Hsm

	for _, c := range out.Clusters {
		for _, h := range c.Hsms {
			if aws.StringValue(h.HsmId) == hsmId {
				hsm = h
				break
			}
		}
	}

	return hsm, nil
}

func resourceAwsCloudHsmV2HsmRefreshFunc(conn *cloudhsmv2.CloudHSMV2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		hsm, err := describeHsm(conn, id)

		if hsm == nil {
			return 42, "destroyed", nil
		}

		if hsm.State != nil {
			log.Printf("[DEBUG] CloudHSMv2 Cluster status (%s): %s", id, *hsm.State)
		}

		return hsm, aws.StringValue(hsm.State), err
	}
}

func resourceAwsCloudHsmV2HsmCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn

	clusterId := d.Get("cluster_id").(string)

	cluster, err := describeCloudHsmV2Cluster(conn, clusterId)

	if cluster == nil {
		log.Printf("[WARN] Error on retrieving CloudHSMv2 Cluster: %s %s", clusterId, err)
		return err
	}

	availabilityZone := d.Get("availability_zone").(string)
	if len(availabilityZone) == 0 {
		subnetId := d.Get("subnet_id").(string)
		for az, sn := range cluster.SubnetMapping {
			if aws.StringValue(sn) == subnetId {
				availabilityZone = az
			}
		}
	}

	input := &cloudhsmv2.CreateHsmInput{
		ClusterId:        aws.String(clusterId),
		AvailabilityZone: aws.String(availabilityZone),
	}

	ipAddress := d.Get("ip_address").(string)
	if len(ipAddress) != 0 {
		input.IpAddress = aws.String(ipAddress)
	}

	log.Printf("[DEBUG] CloudHSMv2 HSM create %s", input)

	var output *cloudhsmv2.CreateHsmOutput

	err = resource.Retry(180*time.Second, func() *resource.RetryError {
		var err error
		output, err = conn.CreateHsm(input)
		if err != nil {
			if isAWSErr(err, cloudhsmv2.ErrCodeCloudHsmInternalFailureException, "request was rejected because of an AWS CloudHSM internal failure") {
				log.Printf("[DEBUG] CloudHSMv2 HSM re-try creating %s", input)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = conn.CreateHsm(input)
	}

	if err != nil {
		return fmt.Errorf("error creating CloudHSM v2 HSM module: %s", err)
	}

	d.SetId(aws.StringValue(output.Hsm.HsmId))

	if err := waitForCloudhsmv2HsmActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for CloudHSMv2 HSM (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsCloudHsmV2HsmRead(d, meta)
}

func resourceAwsCloudHsmV2HsmRead(d *schema.ResourceData, meta interface{}) error {

	hsm, err := describeHsm(meta.(*AWSClient).cloudhsmv2conn, d.Id())

	if hsm == nil {
		log.Printf("[WARN] CloudHSMv2 HSM (%s) not found", d.Id())
		d.SetId("")
		return err
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
	err := resource.Retry(180*time.Second, func() *resource.RetryError {
		var err error
		_, err = conn.DeleteHsm(input)
		if err != nil {
			if isAWSErr(err, cloudhsmv2.ErrCodeCloudHsmInternalFailureException, "request was rejected because of an AWS CloudHSM internal failure") {
				log.Printf("[DEBUG] CloudHSMv2 HSM re-try deleting %s", d.Id())
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteHsm(input)
	}
	if err != nil {
		return fmt.Errorf("error deleting CloudHSM v2 HSM module (%s): %s", d.Id(), err)
	}

	if err := waitForCloudhsmv2HsmDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for CloudHSMv2 HSM (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func waitForCloudhsmv2HsmActive(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateCreateInProgress, "destroyed"},
		Target:     []string{cloudhsmv2.HsmStateActive},
		Refresh:    resourceAwsCloudHsmV2HsmRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForCloudhsmv2HsmDeletion(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateDeleteInProgress},
		Target:     []string{"destroyed"},
		Refresh:    resourceAwsCloudHsmV2HsmRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
