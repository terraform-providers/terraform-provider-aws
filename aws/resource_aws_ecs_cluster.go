package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	ecsClusterTimeoutCreate = 10 * time.Minute
	ecsClusterTimeoutDelete = 10 * time.Minute
	ecsClusterTimeoutUpdate = 10 * time.Minute
)

func resourceAwsEcsCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsClusterCreate,
		Read:   resourceAwsEcsClusterRead,
		Update: resourceAwsEcsClusterUpdate,
		Delete: resourceAwsEcsClusterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEcsClusterImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},

						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},

						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
			"setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ClusterSettingNameContainerInsights,
							}, false),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEcsClusterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Service:   "ecs",
		Resource:  fmt.Sprintf("cluster/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func resourceAwsEcsClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	clusterName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating ECS cluster %s", clusterName)

	input := ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		Tags:        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EcsTags(),
	}

	if v, ok := d.GetOk("setting"); ok {
		input.Settings = expandEcsSettings(v.(*schema.Set))
	}

	if v, ok := d.GetOk("capacity_providers"); ok {
		input.CapacityProviders = expandStringSet(v.(*schema.Set))
	}

	input.DefaultCapacityProviderStrategy = expandEcsCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set))

	out, err := conn.CreateCluster(&input)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] ECS cluster %s created", aws.StringValue(out.Cluster.ClusterArn))

	d.SetId(aws.StringValue(out.Cluster.ClusterArn))

	if err = waitForEcsClusterActive(conn, clusterName, ecsClusterTimeoutCreate); err != nil {
		return fmt.Errorf("error waiting for ECS Cluster (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsEcsClusterRead(d, meta)
}

func resourceAwsEcsClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(d.Id())},
		Include:  []*string{aws.String(ecs.ClusterFieldTags)},
	}

	log.Printf("[DEBUG] Reading ECS Cluster: %s", input)
	var out *ecs.DescribeClustersOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeClusters(input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if out == nil || len(out.Failures) > 0 {
			if d.IsNewResource() {
				return resource.RetryableError(&resource.NotFoundError{})
			}
			return resource.NonRetryableError(&resource.NotFoundError{})
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.DescribeClusters(input)
	}

	if isResourceNotFoundError(err) {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECS Cluster (%s): %s", d.Id(), err)
	}

	var cluster *ecs.Cluster
	for _, c := range out.Clusters {
		if aws.StringValue(c.ClusterArn) == d.Id() {
			cluster = c
			break
		}
	}

	if cluster == nil {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Status==INACTIVE means deleted cluster
	if aws.StringValue(cluster.Status) == "INACTIVE" {
		log.Printf("[WARN] ECS Cluster (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("name", cluster.ClusterName)

	if err := d.Set("capacity_providers", aws.StringValueSlice(cluster.CapacityProviders)); err != nil {
		return fmt.Errorf("error setting capacity_providers: %s", err)
	}
	if err := d.Set("default_capacity_provider_strategy", flattenEcsCapacityProviderStrategy(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return fmt.Errorf("error setting default_capacity_provider_strategy: %s", err)
	}

	if err := d.Set("setting", flattenEcsSettings(cluster.Settings)); err != nil {
		return fmt.Errorf("error setting setting: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.EcsKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsEcsClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	clusterName := d.Get("name").(string)

	if d.HasChange("setting") {
		input := ecs.UpdateClusterSettingsInput{
			Cluster:  aws.String(d.Id()),
			Settings: expandEcsSettings(d.Get("setting").(*schema.Set)),
		}

		_, err := conn.UpdateClusterSettings(&input)
		if err != nil {
			return fmt.Errorf("error changing ECS cluster settings (%s): %s", d.Id(), err)
		}

		if err = waitForEcsClusterActive(conn, clusterName, ecsClusterTimeoutUpdate); err != nil {
			return fmt.Errorf("error waiting for ECS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.EcsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating ECS Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChanges("capacity_providers", "default_capacity_provider_strategy") {
		input := ecs.PutClusterCapacityProvidersInput{
			Cluster:                         aws.String(d.Id()),
			CapacityProviders:               expandStringSet(d.Get("capacity_providers").(*schema.Set)),
			DefaultCapacityProviderStrategy: expandEcsCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set)),
		}

		err := resource.Retry(ecsClusterTimeoutUpdate, func() *resource.RetryError {
			_, err := conn.PutClusterCapacityProviders(&input)
			if err != nil {
				if isAWSErr(err, ecs.ErrCodeClientException, "Cluster was not ACTIVE") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, ecs.ErrCodeResourceInUseException, "") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, ecs.ErrCodeUpdateInProgressException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.PutClusterCapacityProviders(&input)
		}
		if err != nil {
			return fmt.Errorf("error changing ECS cluster capacity provider settings (%s): %s", d.Id(), err)
		}

		if err = waitForEcsClusterActive(conn, clusterName, ecsClusterTimeoutUpdate); err != nil {
			return fmt.Errorf("error waiting for ECS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEcsClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	log.Printf("[DEBUG] Deleting ECS cluster %s", d.Id())
	input := &ecs.DeleteClusterInput{
		Cluster: aws.String(d.Id()),
	}
	err := resource.Retry(ecsClusterTimeoutDelete, func() *resource.RetryError {
		_, err := conn.DeleteCluster(input)

		if err == nil {
			log.Printf("[DEBUG] ECS cluster %s deleted", d.Id())
			return nil
		}

		if isAWSErr(err, "ClusterContainsContainerInstancesException", "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, "ClusterContainsServicesException", "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, ecs.ErrCodeUpdateInProgressException, "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteCluster(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting ECS cluster: %s", err)
	}

	clusterName := d.Get("name").(string)
	stateConf := resource.StateChangeConf{
		Pending: []string{"ACTIVE", "DEPROVISIONING"},
		Target:  []string{"INACTIVE"},
		Timeout: ecsClusterTimeoutDelete,
		Refresh: refreshEcsClusterStatus(conn, clusterName),
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for ECS cluster to become inactive: %s", err)
	}

	log.Printf("[DEBUG] ECS cluster %q deleted", d.Id())
	return nil
}

func waitForEcsClusterActive(conn *ecs.ECS, clusterName string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{"PROVISIONING"},
		Target:  []string{"ACTIVE"},
		Timeout: timeout,
		Refresh: refreshEcsClusterStatus(conn, clusterName),
		Delay:   10 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func refreshEcsClusterStatus(conn *ecs.ECS, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeClusters(&ecs.DescribeClustersInput{
			Clusters: []*string{aws.String(clusterName)},
		})
		if err != nil {
			return nil, "", err
		}
		for _, c := range output.Clusters {
			if aws.StringValue(c.ClusterName) == clusterName {
				return c, aws.StringValue(c.Status), nil
			}
		}
		return nil, "", fmt.Errorf("ECS cluster %q missing", clusterName)
	}
}

func expandEcsSettings(configured *schema.Set) []*ecs.ClusterSetting {
	list := configured.List()
	if len(list) == 0 {
		return nil
	}

	settings := make([]*ecs.ClusterSetting, 0, len(list))

	for _, raw := range list {
		data := raw.(map[string]interface{})

		setting := &ecs.ClusterSetting{
			Name:  aws.String(data["name"].(string)),
			Value: aws.String(data["value"].(string)),
		}

		settings = append(settings, setting)
	}

	return settings
}

func flattenEcsSettings(list []*ecs.ClusterSetting) []map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, setting := range list {
		l := map[string]interface{}{
			"name":  aws.StringValue(setting.Name),
			"value": aws.StringValue(setting.Value),
		}

		result = append(result, l)
	}
	return result
}
