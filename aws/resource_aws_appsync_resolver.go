package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsAppsyncResolver() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppsyncResolverCreate,
		Read:   resourceAwsAppsyncResolverRead,
		Update: resourceAwsAppsyncResolverUpdate,
		Delete: resourceAwsAppsyncResolverDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"field": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data_source": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"pipeline_config"},
			},
			"request_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"response_template": {
				Type:     schema.TypeString,
				Required: true, // documentation bug, the api returns 400 if this is not specified.
			},
			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  appsync.ResolverKindUnit,
				ValidateFunc: validation.StringInSlice([]string{
					appsync.ResolverKindUnit,
					appsync.ResolverKindPipeline,
				}, true),
			},
			"pipeline_config": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"data_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"functions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"caching_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"caching_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"ttl": {
							Type:     schema.TypeInt,
							Optional: true,
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

func resourceAwsAppsyncResolverCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.CreateResolverInput{
		ApiId:                   aws.String(d.Get("api_id").(string)),
		TypeName:                aws.String(d.Get("type").(string)),
		FieldName:               aws.String(d.Get("field").(string)),
		RequestMappingTemplate:  aws.String(d.Get("request_template").(string)),
		ResponseMappingTemplate: aws.String(d.Get("response_template").(string)),
		Kind:                    aws.String(d.Get("kind").(string)),
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_config"); ok {
		config := v.([]interface{})[0].(map[string]interface{})
		input.PipelineConfig = &appsync.PipelineConfig{
			Functions: expandStringList(config["functions"].([]interface{})),
		}
	}

	if v, ok := d.GetOk("caching_config"); ok {
		input.CachingConfig = expandAppsyncResolverCachingConfig(v.([]interface{}))
	}

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	awsMutexKV.Lock(mutexKey)
	defer awsMutexKV.Unlock(mutexKey)

	_, err := retryOnAwsCode(appsync.ErrCodeConcurrentModificationException, func() (interface{}, error) {
		return conn.CreateResolver(input)
	})

	if err != nil {
		return fmt.Errorf("error creating AppSync Resolver: %s", err)
	}

	d.SetId(d.Get("api_id").(string) + "-" + d.Get("type").(string) + "-" + d.Get("field").(string))

	return resourceAwsAppsyncResolverRead(d, meta)
}

func resourceAwsAppsyncResolverRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, typeName, fieldName, err := decodeAppsyncResolverID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.GetResolverInput{
		ApiId:     aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName: aws.String(fieldName),
	}

	resp, err := conn.GetResolver(input)

	if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] AppSync Resolver (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppSync Resolver (%s): %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set("arn", resp.Resolver.ResolverArn)
	d.Set("type", resp.Resolver.TypeName)
	d.Set("field", resp.Resolver.FieldName)
	d.Set("data_source", resp.Resolver.DataSourceName)
	d.Set("request_template", resp.Resolver.RequestMappingTemplate)
	d.Set("response_template", resp.Resolver.ResponseMappingTemplate)
	d.Set("kind", resp.Resolver.Kind)

	if err := d.Set("pipeline_config", flattenAppsyncPipelineConfig(resp.Resolver.PipelineConfig)); err != nil {
		return fmt.Errorf("Error setting pipeline_config: %s", err)
	}

	if err := d.Set("caching_config", flattenAppsyncCachingConfig(resp.Resolver.CachingConfig)); err != nil {
		return fmt.Errorf("Error setting caching_config: %s", err)
	}

	return nil
}

func resourceAwsAppsyncResolverUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.UpdateResolverInput{
		ApiId:                   aws.String(d.Get("api_id").(string)),
		FieldName:               aws.String(d.Get("field").(string)),
		TypeName:                aws.String(d.Get("type").(string)),
		RequestMappingTemplate:  aws.String(d.Get("request_template").(string)),
		ResponseMappingTemplate: aws.String(d.Get("response_template").(string)),
		Kind:                    aws.String(d.Get("kind").(string)),
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSourceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_config"); ok {
		config := v.([]interface{})[0].(map[string]interface{})
		input.PipelineConfig = &appsync.PipelineConfig{
			Functions: expandStringList(config["functions"].([]interface{})),
		}
	}

	if v, ok := d.GetOk("caching_config"); ok {
		input.CachingConfig = expandAppsyncResolverCachingConfig(v.([]interface{}))
	}

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	awsMutexKV.Lock(mutexKey)
	defer awsMutexKV.Unlock(mutexKey)

	_, err := retryOnAwsCode(appsync.ErrCodeConcurrentModificationException, func() (interface{}, error) {
		return conn.UpdateResolver(input)
	})

	if err != nil {
		return fmt.Errorf("error updating AppSync Resolver (%s): %s", d.Id(), err)
	}

	return resourceAwsAppsyncResolverRead(d, meta)
}

func resourceAwsAppsyncResolverDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	apiID, typeName, fieldName, err := decodeAppsyncResolverID(d.Id())

	if err != nil {
		return err
	}

	input := &appsync.DeleteResolverInput{
		ApiId:     aws.String(apiID),
		TypeName:  aws.String(typeName),
		FieldName: aws.String(fieldName),
	}

	mutexKey := fmt.Sprintf("appsync-schema-%s", d.Get("api_id").(string))
	awsMutexKV.Lock(mutexKey)
	defer awsMutexKV.Unlock(mutexKey)

	_, err = retryOnAwsCode(appsync.ErrCodeConcurrentModificationException, func() (interface{}, error) {
		return conn.DeleteResolver(input)
	})

	if err != nil {
		return fmt.Errorf("error deleting AppSync Resolver (%s): %s", d.Id(), err)
	}

	return nil
}

func decodeAppsyncResolverID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "-", 3)
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("expected ID in format ApiID-TypeName-FieldName, received: %s", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func expandAppsyncResolverCachingConfig(l []interface{}) *appsync.CachingConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cachingConfig := &appsync.CachingConfig{
		CachingKeys: expandStringSet(m["caching_keys"].(*schema.Set)),
	}

	if v, ok := m["ttl"].(int); ok && v != 0 {
		cachingConfig.Ttl = aws.Int64(int64(v))
	}

	return cachingConfig
}

func flattenAppsyncPipelineConfig(c *appsync.PipelineConfig) []interface{} {
	if c == nil {
		return nil
	}

	if len(c.Functions) == 0 {
		return nil
	}

	m := map[string]interface{}{
		"functions": flattenStringList(c.Functions),
	}

	return []interface{}{m}
}

func flattenAppsyncCachingConfig(c *appsync.CachingConfig) []interface{} {
	if c == nil {
		return nil
	}

	if len(c.CachingKeys) == 0 && *(c.Ttl) == 0 {
		return nil
	}

	m := map[string]interface{}{
		"caching_keys": flattenStringSet(c.CachingKeys),
		"ttl":          int(aws.Int64Value(c.Ttl)),
	}

	return []interface{}{m}
}
