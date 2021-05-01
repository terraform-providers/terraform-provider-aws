package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsElasticacheSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticacheSubnetGroupCreate,
		Read:   resourceAwsElasticacheSubnetGroupRead,
		Update: resourceAwsElasticacheSubnetGroupUpdate,
		Delete: resourceAwsElasticacheSubnetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					// Elasticache normalizes subnet names to lowercase,
					// so we have to do this too or else we can end up
					// with non-converging diffs.
					return strings.ToLower(val.(string))
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsElasticacheSubnetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	// Get the group properties
	name := d.Get("name").(string)
	desc := d.Get("description").(string)
	subnetIdsSet := d.Get("subnet_ids").(*schema.Set)

	log.Printf("[DEBUG] Cache subnet group create: name: %s, description: %s", name, desc)

	subnetIds := expandStringSet(subnetIdsSet)

	req := &elasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(desc),
		CacheSubnetGroupName:        aws.String(name),
		SubnetIds:                   subnetIds,
		Tags:                        tags.IgnoreAws().ElasticacheTags(),
	}

	_, err := conn.CreateCacheSubnetGroup(req)
	if err != nil {
		return fmt.Errorf("error creating ElastiCache Subnet Group (%s): %w", name, err)
	}

	// Assign the group name as the resource ID
	// Elasticache always retains the name in lower case, so we have to
	// mimic that or else we won't be able to refresh a resource whose
	// name contained uppercase characters.
	d.SetId(strings.ToLower(name))

	return resourceAwsElasticacheSubnetGroupRead(d, meta)
}

func resourceAwsElasticacheSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &elasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: aws.String(d.Get("name").(string)),
	}

	res, err := conn.DescribeCacheSubnetGroups(req)
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "CacheSubnetGroupNotFoundFault" {
			// Update state to indicate the db subnet no longer exists.
			log.Printf("[WARN] Elasticache Subnet Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if len(res.CacheSubnetGroups) == 0 {
		return fmt.Errorf("Error missing %v", d.Get("name"))
	}

	var group *elasticache.CacheSubnetGroup
	for _, g := range res.CacheSubnetGroups {
		log.Printf("[DEBUG] %v %v", g.CacheSubnetGroupName, d.Id())
		if *g.CacheSubnetGroupName == d.Id() {
			group = g
		}
	}
	if group == nil {
		return fmt.Errorf("Error retrieving cache subnet group: %v", res)
	}

	ids := make([]string, len(group.Subnets))
	for i, s := range group.Subnets {
		ids[i] = *s.SubnetIdentifier
	}

	d.Set("arn", group.ARN)
	d.Set("name", group.CacheSubnetGroupName)
	d.Set("description", group.CacheSubnetGroupDescription)
	d.Set("subnet_ids", ids)

	tags, err := keyvaluetags.ElasticacheListTags(conn, d.Get("arn").(string))

	if err != nil && !isAWSErr(err, "UnknownOperationException", "") {
		return fmt.Errorf("error listing tags for ElastiCache SubnetGroup (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsElasticacheSubnetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	if d.HasChanges("subnet_ids", "description") {
		var subnets []*string
		if v := d.Get("subnet_ids"); v != nil {
			for _, v := range v.(*schema.Set).List() {
				subnets = append(subnets, aws.String(v.(string)))
			}
		}
		log.Printf("[DEBUG] Updating ElastiCache Subnet Group")

		_, err := conn.ModifyCacheSubnetGroup(&elasticache.ModifyCacheSubnetGroupInput{
			CacheSubnetGroupName:        aws.String(d.Get("name").(string)),
			CacheSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                   subnets,
		})

		if err != nil {
			return fmt.Errorf("error updating ElastiCache Subnet Group (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.ElasticacheUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsElasticacheSubnetGroupRead(d, meta)
}
func resourceAwsElasticacheSubnetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticacheconn

	log.Printf("[DEBUG] Cache subnet group delete: %s", d.Id())

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteCacheSubnetGroup(&elasticache.DeleteCacheSubnetGroupInput{
			CacheSubnetGroupName: aws.String(d.Id()),
		})
		if err != nil {
			apierr, ok := err.(awserr.Error)
			if !ok {
				return resource.RetryableError(err)
			}
			log.Printf("[DEBUG] APIError.Code: %v", apierr.Code())
			switch apierr.Code() {
			case "DependencyViolation":
				// If it is a dependency violation, we want to retry
				return resource.RetryableError(err)
			default:
				return resource.NonRetryableError(err)
			}
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteCacheSubnetGroup(&elasticache.DeleteCacheSubnetGroupInput{
			CacheSubnetGroupName: aws.String(d.Id()),
		})
	}

	if isAWSErr(err, elasticache.ErrCodeCacheSubnetGroupNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Subnet Group (%s): %w", d.Id(), err)
	}

	return nil
}
