package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsFmsPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFmsPolicyCreate,
		Read:   resourceAwsFmsPolicyRead,
		Update: resourceAwsFmsPolicyUpdate,
		Delete: resourceAwsFmsPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"delete_all_policy_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"exclude_resource_tags": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"exclude_map": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"include_map": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"orgunit": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"remediation_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"resource_type_list": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Set:           schema.HashString,
				ConflictsWith: []string{"resource_type"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"AWS::ApiGateway::Stage",
						"AWS::CloudFront::Distribution",
						"AWS::EC2::NetworkInterface",
						"AWS::EC2::Instance",
						"AWS::EC2::SecurityGroup",
						"AWS::EC2::VPC",
						"AWS::ElasticLoadBalancingV2::LoadBalancer",
					}, false),
				},
			},

			"resource_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"resource_type_list"},
			},

			"policy_update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_tags": tagsSchema(),

			"security_service_policy_data": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"managed_service_data": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressEquivalentJsonDiffs,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsFmsPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	fmsPolicy := resourceAwsFmsPolicyExpandPolicy(d)

	params := &fms.PutPolicyInput{
		Policy: fmsPolicy,
	}

	var resp *fms.PutPolicyOutput
	var err error

	resp, err = conn.PutPolicy(params)

	if err != nil {
		return fmt.Errorf("Creating Policy Failed: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.Policy.PolicyId))

	return resourceAwsFmsPolicyRead(d, meta)
}

func resourceAwsFmsPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	var resp *fms.GetPolicyOutput
	var req = &fms.GetPolicyInput{
		PolicyId: aws.String(d.Id()),
	}

	resp, err := conn.GetPolicy(req)

	if err != nil {
		if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] FMS Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	return resourceAwsFmsPolicyFlattenPolicy(d, resp)
}

func resourceAwsFmsPolicyFlattenPolicy(d *schema.ResourceData, resp *fms.GetPolicyOutput) error {
	d.Set("arn", aws.StringValue(resp.PolicyArn))

	d.Set("name", aws.StringValue(resp.Policy.PolicyName))
	d.Set("exclude_resource_tags", aws.BoolValue(resp.Policy.ExcludeResourceTags))
	if err := d.Set("exclude_map", flattenFMSPolicyMap(resp.Policy.ExcludeMap)); err != nil {
		return err
	}
	if err := d.Set("include_map", flattenFMSPolicyMap(resp.Policy.IncludeMap)); err != nil {
		return err
	}
	d.Set("remediation_enabled", aws.BoolValue(resp.Policy.RemediationEnabled))
	if err := d.Set("resource_type_list", resp.Policy.ResourceTypeList); err != nil {
		return err
	}
	d.Set("resource_type", aws.StringValue(resp.Policy.ResourceType))
	d.Set("policy_update_token", aws.StringValue(resp.Policy.PolicyUpdateToken))
	if err := d.Set("resource_tags", flattenFMSResourceTags(resp.Policy.ResourceTags)); err != nil {
		return err
	}

	securityServicePolicy := []map[string]string{{
		"type":                 *resp.Policy.SecurityServicePolicyData.Type,
		"managed_service_data": *resp.Policy.SecurityServicePolicyData.ManagedServiceData,
	}}
	if err := d.Set("security_service_policy_data", securityServicePolicy); err != nil {
		return err
	}

	return nil
}

func resourceAwsFmsPolicyExpandPolicy(d *schema.ResourceData) *fms.Policy {
	resourceType := aws.String("ResourceTypeList")
	resourceTypeList := expandStringSet(d.Get("resource_type_list").(*schema.Set))
	if t, ok := d.GetOk("resource_type"); ok {
		resourceType = aws.String(t.(string))
	}

	fmsPolicy := &fms.Policy{
		PolicyName:          aws.String(d.Get("name").(string)),
		RemediationEnabled:  aws.Bool(d.Get("remediation_enabled").(bool)),
		ResourceType:        resourceType,
		ResourceTypeList:    resourceTypeList,
		ExcludeResourceTags: aws.Bool(d.Get("exclude_resource_tags").(bool)),
	}

	if d.Id() != "" {
		fmsPolicy.PolicyId = aws.String(d.Id())
		fmsPolicy.PolicyUpdateToken = aws.String(d.Get("policy_update_token").(string))
	}

	fmsPolicy.ExcludeMap = expandFMSPolicyMap(d.Get("exclude_map").([]interface{}))

	fmsPolicy.IncludeMap = expandFMSPolicyMap(d.Get("include_map").([]interface{}))

	fmsPolicy.ResourceTags = constructResourceTags(d.Get("resource_tags"))

	securityServicePolicy := d.Get("security_service_policy_data").([]interface{})[0].(map[string]interface{})
	fmsPolicy.SecurityServicePolicyData = &fms.SecurityServicePolicyData{
		ManagedServiceData: aws.String(securityServicePolicy["managed_service_data"].(string)),
		Type:               aws.String(securityServicePolicy["type"].(string)),
	}

	return fmsPolicy
}

func resourceAwsFmsPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	fmsPolicy := resourceAwsFmsPolicyExpandPolicy(d)

	params := &fms.PutPolicyInput{Policy: fmsPolicy}
	_, err := conn.PutPolicy(params)

	if err != nil {
		return fmt.Errorf("Error modifying FMS Policy Rule: %s", err)
	}

	return resourceAwsFmsPolicyRead(d, meta)
}

func resourceAwsFmsPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn
	log.Printf("[DEBUG] Delete FMS Policy: %s", d.Id())

	_, err := conn.DeletePolicy(&fms.DeletePolicyInput{
		PolicyId:                 aws.String(d.Id()),
		DeleteAllPolicyResources: aws.Bool(d.Get("delete_all_policy_resources").(bool)),
	})

	if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FMS Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func expandFMSPolicyMap(set []interface{}) map[string][]*string {
	fmsPolicyMap := map[string][]*string{}
	if len(set) > 0 {
		if _, ok := set[0].(map[string]interface{}); !ok {
			return fmsPolicyMap
		}
		for key, listValue := range set[0].(map[string]interface{}) {
			var flatKey string
			switch key {
			case "account":
				flatKey = "ACCOUNT"
			case "orgunit":
				flatKey = "ORG_UNIT"
			}

			for _, value := range listValue.(*schema.Set).List() {
				fmsPolicyMap[flatKey] = append(fmsPolicyMap[flatKey], aws.String(value.(string)))
			}
		}
	}
	return fmsPolicyMap
}

func flattenFMSPolicyMap(fmsPolicyMap map[string][]*string) []interface{} {
	flatPolicyMap := map[string]interface{}{}

	for key, value := range fmsPolicyMap {
		switch key {
		case "ACCOUNT":
			flatPolicyMap["account"] = value
		case "ORG_UNIT":
			flatPolicyMap["orgunit"] = value
		default:
			log.Printf("[WARNING] Unexpected key (%q) found in FMS policy", key)
		}
	}

	return []interface{}{flatPolicyMap}
}

func flattenFMSResourceTags(resourceTags []*fms.ResourceTag) map[string]interface{} {
	resTags := map[string]interface{}{}

	for _, v := range resourceTags {
		resTags[*v.Key] = v.Value
	}
	return resTags
}

func constructResourceTags(rTags interface{}) []*fms.ResourceTag {
	var rTagList []*fms.ResourceTag

	tags := rTags.(map[string]interface{})
	for k, v := range tags {
		rTagList = append(rTagList, &fms.ResourceTag{Key: aws.String(k), Value: aws.String(v.(string))})
	}

	return rTagList
}
