package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCloudHsmV2Cluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudHsmV2ClusterCreate,
		Read:   resourceAwsCloudHsmV2ClusterRead,
		Update: resourceAwsCloudHsmV2ClusterUpdate,
		Delete: resourceAwsCloudHsmV2ClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"source_backup_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"hsm_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"hsm1.medium"}, false),
			},

			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				ForceNew: true,
			},

			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cluster_certificates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_csr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"aws_hardware_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hsm_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"manufacturer_hardware_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cluster_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func describeCloudHsmV2Cluster(conn *cloudhsmv2.CloudHSMV2, clusterId string) (*cloudhsmv2.Cluster, error) {
	filters := []*string{&clusterId}
	result := int64(1)
	out, err := conn.DescribeClusters(&cloudhsmv2.DescribeClustersInput{
		Filters: map[string][]*string{
			"clusterIds": filters,
		},
		MaxResults: &result,
	})
	if err != nil {
		log.Printf("[WARN] Error on retrieving CloudHSMv2 Cluster (%s) when waiting: %s", clusterId, err)
		return nil, err
	}

	var cluster *cloudhsmv2.Cluster

	for _, c := range out.Clusters {
		if aws.StringValue(c.ClusterId) == clusterId {
			cluster = c
			break
		}
	}
	return cluster, nil
}

func resourceAwsCloudHsmV2ClusterRefreshFunc(conn *cloudhsmv2.CloudHSMV2, clusterId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := describeCloudHsmV2Cluster(conn, clusterId)

		if cluster == nil {
			return 42, "destroyed", nil
		}

		if cluster.State != nil {
			log.Printf("[DEBUG] CloudHSMv2 Cluster status (%s): %s", clusterId, *cluster.State)
		}

		return cluster, aws.StringValue(cluster.State), err
	}
}

func resourceAwsCloudHsmV2ClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn

	input := &cloudhsmv2.CreateClusterInput{
		HsmType:   aws.String(d.Get("hsm_type").(string)),
		SubnetIds: expandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.TagList = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Cloudhsmv2Tags()
	}

	backupId := d.Get("source_backup_identifier").(string)
	if len(backupId) != 0 {
		input.SourceBackupId = aws.String(backupId)
	}

	log.Printf("[DEBUG] CloudHSMv2 Cluster create %s", input)

	var output *cloudhsmv2.CreateClusterOutput

	err := resource.Retry(180*time.Second, func() *resource.RetryError {
		var err error
		output, err = conn.CreateCluster(input)
		if err != nil {
			if isAWSErr(err, cloudhsmv2.ErrCodeCloudHsmInternalFailureException, "request was rejected because of an AWS CloudHSM internal failure") {
				log.Printf("[DEBUG] CloudHSMv2 Cluster re-try creating %s", input)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = conn.CreateCluster(input)
	}

	if err != nil {
		return fmt.Errorf("error creating CloudHSMv2 Cluster: %s", err)
	}

	d.SetId(aws.StringValue(output.Cluster.ClusterId))
	log.Printf("[INFO] CloudHSMv2 Cluster ID: %s", d.Id())
	log.Println("[INFO] Waiting for CloudHSMv2 Cluster to be available")

	targetState := cloudhsmv2.ClusterStateUninitialized
	if len(backupId) > 0 {
		targetState = cloudhsmv2.ClusterStateActive
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.ClusterStateCreateInProgress, cloudhsmv2.ClusterStateInitializeInProgress},
		Target:     []string{targetState},
		Refresh:    resourceAwsCloudHsmV2ClusterRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, errWait := stateConf.WaitForState()
	if errWait != nil {
		if len(backupId) == 0 {
			return fmt.Errorf("[WARN] Error waiting for CloudHSMv2 Cluster state to be \"UNINITIALIZED\": %s", errWait)
		} else {
			return fmt.Errorf("[WARN] Error waiting for CloudHSMv2 Cluster state to be \"ACTIVE\": %s", errWait)
		}
	}

	return resourceAwsCloudHsmV2ClusterRead(d, meta)
}

func resourceAwsCloudHsmV2ClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	cluster, err := describeCloudHsmV2Cluster(conn, d.Id())

	if cluster == nil {
		log.Printf("[WARN] CloudHSMv2 Cluster (%s) not found", d.Id())
		d.SetId("")
		return err
	}

	log.Printf("[INFO] Reading CloudHSMv2 Cluster Information: %s", d.Id())

	d.Set("cluster_id", cluster.ClusterId)
	d.Set("cluster_state", cluster.State)
	d.Set("security_group_id", cluster.SecurityGroup)
	d.Set("vpc_id", cluster.VpcId)
	d.Set("source_backup_identifier", cluster.SourceBackupId)
	d.Set("hsm_type", cluster.HsmType)
	if err := d.Set("cluster_certificates", readCloudHsmV2ClusterCertificates(cluster)); err != nil {
		return fmt.Errorf("error setting cluster_certificates: %s", err)
	}

	var subnets []string
	for _, sn := range cluster.SubnetMapping {
		subnets = append(subnets, aws.StringValue(sn))
	}
	if err := d.Set("subnet_ids", subnets); err != nil {
		return fmt.Errorf("Error saving Subnet IDs to state for CloudHSMv2 Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", keyvaluetags.Cloudhsmv2KeyValueTags(cluster.TagList).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsCloudHsmV2ClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Cloudhsmv2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsCloudHsmV2ClusterRead(d, meta)
}

func resourceAwsCloudHsmV2ClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn
	input := &cloudhsmv2.DeleteClusterInput{
		ClusterId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] CloudHSMv2 Delete cluster: %s", d.Id())
	err := resource.Retry(180*time.Second, func() *resource.RetryError {
		var err error
		_, err = conn.DeleteCluster(input)
		if err != nil {
			if isAWSErr(err, cloudhsmv2.ErrCodeCloudHsmInternalFailureException, "request was rejected because of an AWS CloudHSM internal failure") {
				log.Printf("[DEBUG] CloudHSMv2 Cluster re-try deleting %s", d.Id())
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteCluster(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudHSMv2 Cluster (%s): %s", d.Id(), err)
	}

	if err := waitForCloudhsmv2ClusterDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for CloudHSMv2 Cluster (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func readCloudHsmV2ClusterCertificates(cluster *cloudhsmv2.Cluster) []map[string]interface{} {
	certs := map[string]interface{}{}
	if cluster.Certificates != nil {
		if aws.StringValue(cluster.State) == cloudhsmv2.ClusterStateUninitialized {
			certs["cluster_csr"] = aws.StringValue(cluster.Certificates.ClusterCsr)
			certs["aws_hardware_certificate"] = aws.StringValue(cluster.Certificates.AwsHardwareCertificate)
			certs["hsm_certificate"] = aws.StringValue(cluster.Certificates.HsmCertificate)
			certs["manufacturer_hardware_certificate"] = aws.StringValue(cluster.Certificates.ManufacturerHardwareCertificate)
		} else if aws.StringValue(cluster.State) == cloudhsmv2.ClusterStateActive {
			certs["cluster_certificate"] = aws.StringValue(cluster.Certificates.ClusterCertificate)
		}
	}
	if len(certs) > 0 {
		return []map[string]interface{}{certs}
	}
	return []map[string]interface{}{}
}

func waitForCloudhsmv2ClusterDeletion(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.ClusterStateDeleteInProgress},
		Target:     []string{cloudhsmv2.ClusterStateDeleted},
		Refresh:    resourceAwsCloudHsmV2ClusterRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
