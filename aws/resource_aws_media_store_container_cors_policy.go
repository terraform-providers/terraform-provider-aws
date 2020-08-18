package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsMediaStoreContainerCorsPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaStoreContainerCorsPolicyPut,
		Read:   resourceAwsMediaStoreContainerCorsPolicyRead,
		Update: resourceAwsMediaStoreContainerCorsPolicyPut,
		Delete: resourceAwsMediaStoreContainerCorsPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cors_policy": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 100,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"allowed_methods": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 4,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									mediastore.MethodNamePut,
									mediastore.MethodNameGet,
									mediastore.MethodNameDelete,
									mediastore.MethodNameHead,
								}, false),
							},
							Set: schema.HashString,
						},
						"allowed_origins": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 100,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 100,
							MinItems: 0,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"max_age_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 2147483647),
						},
					},
				},
			},
		},
	}
}

func resourceAwsMediaStoreContainerCorsPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.PutCorsPolicyInput{
		ContainerName: aws.String(d.Get("container_name").(string)),
		CorsPolicy:    expandMediaStoreContainerCorsPolicy(d.Get("cors_policy").([]interface{})),
	}

	log.Printf("[DEBUG] Media Store Container Cors Policy put configuration: %#v", input)
	_, err := conn.PutCorsPolicy(input)
	if err != nil {
		return fmt.Errorf("Error putting cors policy: %s", err)
	}

	d.SetId(d.Get("container_name").(string))

	return resourceAwsMediaStoreContainerCorsPolicyRead(d, meta)
}

func resourceAwsMediaStoreContainerCorsPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.GetCorsPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Media Store Container Cors Policy describe configuration: %#v", input)
	resp, err := conn.GetCorsPolicy(input)
	if isAWSErr(err, mediastore.ErrCodeCorsPolicyNotFoundException, "") || isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
		log.Printf("[WARN] Media Store Container Cors Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting cors policy: %s", err)
	}

	d.Set("container_name", d.Id())
	d.Set("cors_policy", flattenMediaStoreContainerCorsPolicy(resp.CorsPolicy))

	return nil
}

func resourceAwsMediaStoreContainerCorsPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.DeleteCorsPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Media Store Container Cors Policy delete configuration: %#v", input)
	_, err := conn.DeleteCorsPolicy(input)
	if isAWSErr(err, mediastore.ErrCodeCorsPolicyNotFoundException, "") || isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting cors policy: %s", err)
	}

	return nil
}

func expandMediaStoreContainerCorsPolicy(rawCorsRules []interface{}) []*mediastore.CorsRule {
	corsRules := make([]*mediastore.CorsRule, 0, len(rawCorsRules))
	for _, rawCorsRule := range rawCorsRules {
		cr := rawCorsRule.(map[string]interface{})
		corsRule := &mediastore.CorsRule{}
		if v, ok := cr["allowed_headers"]; ok {
			corsRule.AllowedHeaders = expandStringSet(v.(*schema.Set))
		}
		if v, ok := cr["allowed_methods"]; ok {
			corsRule.AllowedMethods = expandStringSet(v.(*schema.Set))
		}
		if v, ok := cr["allowed_origins"]; ok {
			corsRule.AllowedOrigins = expandStringSet(v.(*schema.Set))
		}
		if v, ok := cr["expose_headers"]; ok && len(v.(*schema.Set).List()) > 0 {
			corsRule.ExposeHeaders = expandStringSet(v.(*schema.Set))
		}
		if v, ok := cr["max_age_seconds"]; ok {
			corsRule.MaxAgeSeconds = aws.Int64(int64(v.(int)))
		}
		corsRules = append(corsRules, corsRule)
	}
	return corsRules
}

func flattenMediaStoreContainerCorsPolicy(corsRules []*mediastore.CorsRule) []interface{} {
	rawCorsRules := make([]interface{}, 0, len(corsRules))

	for _, corsRule := range corsRules {
		m := map[string]interface{}{
			"allowed_headers": schema.NewSet(schema.HashString, flattenStringList(corsRule.AllowedHeaders)),
			"allowed_methods": schema.NewSet(schema.HashString, flattenStringList(corsRule.AllowedMethods)),
			"allowed_origins": schema.NewSet(schema.HashString, flattenStringList(corsRule.AllowedOrigins)),
			"expose_headers":  schema.NewSet(schema.HashString, flattenStringList(corsRule.ExposeHeaders)),
			"max_age_seconds": aws.Int64Value(corsRule.MaxAgeSeconds),
		}
		rawCorsRules = append(rawCorsRules, m)
	}

	return rawCorsRules
}
