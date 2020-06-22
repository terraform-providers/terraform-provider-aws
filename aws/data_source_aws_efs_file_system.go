package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEfsFileSystem() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEfsFileSystemRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"throughput_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provisioned_throughput_in_mibps": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"size_in_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"lifecycle_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transition_to_ia": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsEfsFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	efsconn := meta.(*AWSClient).efsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeEfsOpts := &efs.DescribeFileSystemsInput{}

	if v, ok := d.GetOk("creation_token"); ok {
		describeEfsOpts.CreationToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_system_id"); ok {
		describeEfsOpts.FileSystemId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Reading EFS File System: %s", describeEfsOpts)
	describeResp, err := efsconn.DescribeFileSystems(describeEfsOpts)
	if err != nil {
		return fmt.Errorf("Error retrieving EFS: %s", err)
	}
	if len(describeResp.FileSystems) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(describeResp.FileSystems))
	}

	d.SetId(*describeResp.FileSystems[0].FileSystemId)

	var fs *efs.FileSystemDescription
	for _, f := range describeResp.FileSystems {
		if d.Id() == *f.FileSystemId {
			fs = f
			break
		}
	}
	if fs == nil {
		log.Printf("[WARN] EFS (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("creation_token", fs.CreationToken)
	d.Set("performance_mode", fs.PerformanceMode)

	fsARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(fs.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()

	d.Set("arn", fsARN)
	d.Set("file_system_id", fs.FileSystemId)
	d.Set("encrypted", fs.Encrypted)
	d.Set("kms_key_id", fs.KmsKeyId)
	d.Set("provisioned_throughput_in_mibps", fs.ProvisionedThroughputInMibps)
	d.Set("throughput_mode", fs.ThroughputMode)
	if fs.SizeInBytes != nil {
		d.Set("size_in_bytes", fs.SizeInBytes.Value)
	}

	if err := d.Set("tags", keyvaluetags.EfsKeyValueTags(fs.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	res, err := efsconn.DescribeLifecycleConfiguration(&efs.DescribeLifecycleConfigurationInput{
		FileSystemId: fs.FileSystemId,
	})
	if err != nil {
		return fmt.Errorf("Error describing lifecycle configuration for EFS file system (%s): %s",
			aws.StringValue(fs.FileSystemId), err)
	}

	if err := d.Set("lifecycle_policy", flattenEfsFileSystemLifecyclePolicies(res.LifecyclePolicies)); err != nil {
		return fmt.Errorf("error setting lifecycle_policy: %s", err)
	}

	d.Set("dns_name", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(fs.FileSystemId))))

	return nil
}
