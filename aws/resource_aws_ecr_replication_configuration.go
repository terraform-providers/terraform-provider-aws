package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsEcrReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcrReplicationConfigurationPut,
		Read:   resourceAwsEcrReplicationConfigurationRead,
		Update: resourceAwsEcrReplicationConfigurationPut,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"region": {
													Type:     schema.TypeString,
													Required: true,
												},
												"registry_id": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateAwsAccountId,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsEcrReplicationConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	input := ecr.PutReplicationConfigurationInput{
		ReplicationConfiguration: expandEcrReplicationConfigurationReplicationConfiguration(d.Get("replication_configuration").([]interface{})),
	}

	_, err := conn.PutReplicationConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error creating ECR Replication Configuration: %w", err)
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceAwsEcrReplicationConfigurationRead(d, meta)
}

func resourceAwsEcrReplicationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	log.Printf("[DEBUG] Reading ECR Replication Configuration %s", d.Id())
	out, err := conn.DescribeRegistry(&ecr.DescribeRegistryInput{})
	if err != nil {
		return fmt.Errorf("error reading ECR Replication Configuration: %w", err)
	}

	d.Set("registry_id", out.RegistryId)

	if err := d.Set("replication_configuration", flattenEcrReplicationConfigurationReplicationConfiguration(out.ReplicationConfiguration)); err != nil {
		return fmt.Errorf("error setting replication_configuration for ECR Replication Configuration: %w", err)
	}

	return nil
}

func expandEcrReplicationConfigurationReplicationConfiguration(data []interface{}) *ecr.ReplicationConfiguration {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	ec := data[0].(map[string]interface{})
	config := &ecr.ReplicationConfiguration{
		Rules: expandEcrReplicationConfigurationReplicationConfigurationRules(ec["rule"].([]interface{})),
	}
	return config
}

func flattenEcrReplicationConfigurationReplicationConfiguration(ec *ecr.ReplicationConfiguration) []map[string]interface{} {
	if ec == nil {
		return nil
	}

	config := map[string]interface{}{
		"rule": flattenEcrReplicationConfigurationReplicationConfigurationRules(ec.Rules),
	}

	return []map[string]interface{}{
		config,
	}
}

func expandEcrReplicationConfigurationReplicationConfigurationRules(data []interface{}) []*ecr.ReplicationRule {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var rules []*ecr.ReplicationRule

	for _, rule := range data {
		ec := rule.(map[string]interface{})
		config := &ecr.ReplicationRule{
			Destinations: expandEcrReplicationConfigurationReplicationConfigurationRulesDestinations(ec["destination"].([]interface{})),
		}

		rules = append(rules, config)

	}
	return rules
}

func flattenEcrReplicationConfigurationReplicationConfigurationRules(ec []*ecr.ReplicationRule) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			"destination": flattenEcrReplicationConfigurationReplicationConfigurationRulesDestinations(apiObject.Destinations),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandEcrReplicationConfigurationReplicationConfigurationRulesDestinations(data []interface{}) []*ecr.ReplicationDestination {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var dests []*ecr.ReplicationDestination

	for _, dest := range data {
		ec := dest.(map[string]interface{})
		config := &ecr.ReplicationDestination{
			Region:     aws.String(ec["region"].(string)),
			RegistryId: aws.String(ec["registry_id"].(string)),
		}

		dests = append(dests, config)
	}
	return dests
}

func flattenEcrReplicationConfigurationReplicationConfigurationRulesDestinations(ec []*ecr.ReplicationDestination) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			"region":      aws.StringValue(apiObject.Region),
			"registry_id": aws.StringValue(apiObject.RegistryId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
