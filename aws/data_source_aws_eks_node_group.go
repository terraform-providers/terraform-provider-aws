package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEksNodeGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksNodeGroupRead,

		Schema: map[string]*schema.Schema{
			"ami_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"instance_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"node_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"node_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"release_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remote_access": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ec2_ssh_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoscaling_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"remote_access_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"scaling_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"min_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEksNodeGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	clusterName := d.Get("cluster_name").(string)
	nodeGroupName := d.Get("node_group_name").(string)

	input := &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	}

	log.Printf("[DEBUG] Reading EKS Node Group: %s", input)
	output, err := conn.DescribeNodegroup(input)
	if err != nil {
		return err
	}

	if output == nil || output.Nodegroup == nil {
		return fmt.Errorf("EKS Node Group (%s) not found", nodeGroupName)
	}

	if aws.StringValue(output.Nodegroup.NodegroupName) != nodeGroupName {
		return fmt.Errorf("EKS Node Group (%s) not found", nodeGroupName)
	}

	nodeGroup := output.Nodegroup

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(nodeGroup.ClusterName), aws.StringValue(nodeGroup.NodegroupName)))

	d.Set("ami_type", nodeGroup.AmiType)
	d.Set("arn", nodeGroup.NodegroupArn)
	d.Set("cluster_name", nodeGroup.ClusterName)
	d.Set("disk_size", nodeGroup.DiskSize)
	d.Set("instance_types", nodeGroup.InstanceTypes)
	d.Set("labels", nodeGroup.Labels)
	d.Set("node_group_name", nodeGroup.NodegroupName)
	d.Set("node_role_arn", nodeGroup.NodeRole)
	d.Set("release_version", nodeGroup.ReleaseVersion)

	if err := d.Set("remote_access", flattenEksRemoteAccessConfig(nodeGroup.RemoteAccess)); err != nil {
		return fmt.Errorf("error setting remote_access: %s", err)
	}

	if err := d.Set("resources", flattenEksNodeGroupResources(nodeGroup.Resources)); err != nil {
		return fmt.Errorf("error setting resources: %s", err)
	}

	if err := d.Set("scaling_config", flattenEksNodeGroupScalingConfig(nodeGroup.ScalingConfig)); err != nil {
		return fmt.Errorf("error setting scaling_config: %s", err)
	}

	d.Set("status", nodeGroup.Status)

	if err := d.Set("subnet_ids", aws.StringValueSlice(nodeGroup.Subnets)); err != nil {
		return fmt.Errorf("error setting subnets: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(nodeGroup.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("version", nodeGroup.Version)

	return nil
}
