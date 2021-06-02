package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudhsmv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudhsmv2/waiter"
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

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsCloudHsmV2ClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &cloudhsmv2.CreateClusterInput{
		HsmType:   aws.String(d.Get("hsm_type").(string)),
		SubnetIds: expandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.TagList = tags.IgnoreAws().Cloudhsmv2Tags()
	}

	if v, ok := d.GetOk("source_backup_identifier"); ok {
		input.SourceBackupId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] CloudHSMv2 Cluster create %s", input)

	output, err := conn.CreateCluster(input)

	if err != nil {
		return fmt.Errorf("error creating CloudHSMv2 Cluster: %w", err)
	}

	d.SetId(aws.StringValue(output.Cluster.ClusterId))
	log.Printf("[INFO] CloudHSMv2 Cluster ID: %s", d.Id())
	log.Println("[INFO] Waiting for CloudHSMv2 Cluster to be available")

	if input.SourceBackupId != nil {
		if _, err := waiter.ClusterActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("error waiting for CloudHSMv2 Cluster (%s) creation: %w", d.Id(), err)
		}
	} else {
		if _, err := waiter.ClusterUninitialized(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("error waiting for CloudHSMv2 Cluster (%s) creation: %w", d.Id(), err)
		}
	}

	return resourceAwsCloudHsmV2ClusterRead(d, meta)
}

func resourceAwsCloudHsmV2ClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	cluster, err := finder.Cluster(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading CloudHSMv2 Cluster (%s): %w", d.Id(), err)
	}

	if cluster == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading CloudHSMv2 Cluster (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] CloudHSMv2 Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(cluster.State) == cloudhsmv2.ClusterStateDeleted {
		if d.IsNewResource() {
			return fmt.Errorf("error reading CloudHSMv2 Cluster (%s): %s after creation", d.Id(), aws.StringValue(cluster.State))
		}

		log.Printf("[WARN] CloudHSMv2 Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
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

	tags := keyvaluetags.Cloudhsmv2KeyValueTags(cluster.TagList).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsCloudHsmV2ClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudhsmv2conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
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

	_, err := conn.DeleteCluster(input)

	if err != nil {
		return fmt.Errorf("error deleting CloudHSMv2 Cluster (%s): %w", d.Id(), err)
	}

	if _, err := waiter.ClusterDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for CloudHSMv2 Cluster (%s) deletion: %w", d.Id(), err)
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
