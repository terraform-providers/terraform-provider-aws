package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsS3BucketCorsConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketCorsConfigurationUpdate,
		Read:   resourceAwsS3BucketCorsConfigurationRead,
		Update: resourceAwsS3BucketCorsConfigurationUpdate,
		Delete: resourceAwsS3BucketCorsConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},

			"cors_rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_origins": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsS3BucketCorsConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	corsResponse, err := retryOnAwsCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return s3conn.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: aws.String(bucket),
		})
	})
	if err != nil && !isAWSErr(err, "NoSuchCORSConfiguration", "") {
		d.SetId("")
		return fmt.Errorf("error getting S3 Bucket CORS configuration: %s", err)
	}

	corsRules := make([]map[string]interface{}, 0)
	if cors, ok := corsResponse.(*s3.GetBucketCorsOutput); ok && len(cors.CORSRules) > 0 {
		corsRules = make([]map[string]interface{}, 0, len(cors.CORSRules))
		for _, ruleObject := range cors.CORSRules {
			rule := make(map[string]interface{})
			rule["allowed_headers"] = flattenStringList(ruleObject.AllowedHeaders)
			rule["allowed_methods"] = flattenStringList(ruleObject.AllowedMethods)
			rule["allowed_origins"] = flattenStringList(ruleObject.AllowedOrigins)
			// Both the "ExposeHeaders" and "MaxAgeSeconds" might not be set.
			if ruleObject.AllowedOrigins != nil {
				rule["expose_headers"] = flattenStringList(ruleObject.ExposeHeaders)
			}
			if ruleObject.MaxAgeSeconds != nil {
				rule["max_age_seconds"] = int(*ruleObject.MaxAgeSeconds)
			}
			corsRules = append(corsRules, rule)
		}
	}
	if err := d.Set("cors_rule", corsRules); err != nil {
		return fmt.Errorf("error setting cors_rule: %s", err)
	}

	return nil
}

func resourceAwsS3BucketCorsConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	rawCors := d.Get("cors_rule").([]interface{})

	if len(rawCors) != 0 {
		// Put CORS
		rules := make([]*s3.CORSRule, 0, len(rawCors))
		for _, cors := range rawCors {
			corsMap := cors.(map[string]interface{})
			r := &s3.CORSRule{}
			for k, v := range corsMap {
				log.Printf("[DEBUG] S3 bucket: %s, put CORS: %#v, %#v", bucket, k, v)
				if k == "max_age_seconds" {
					r.MaxAgeSeconds = aws.Int64(int64(v.(int)))
				} else {
					vMap := make([]*string, len(v.([]interface{})))
					for i, vv := range v.([]interface{}) {
						if str, ok := vv.(string); ok {
							vMap[i] = aws.String(str)
						}
					}
					switch k {
					case "allowed_headers":
						r.AllowedHeaders = vMap
					case "allowed_methods":
						r.AllowedMethods = vMap
					case "allowed_origins":
						r.AllowedOrigins = vMap
					case "expose_headers":
						r.ExposeHeaders = vMap
					}
				}
			}
			rules = append(rules, r)
		}
		corsInput := &s3.PutBucketCorsInput{
			Bucket: aws.String(bucket),
			CORSConfiguration: &s3.CORSConfiguration{
				CORSRules: rules,
			},
		}
		log.Printf("[DEBUG] S3 bucket: %s, put CORS: %#v", bucket, corsInput)

		_, err := retryOnAwsCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
			return s3conn.PutBucketCors(corsInput)
		})
		if err != nil {
			return fmt.Errorf("Error putting S3 CORS: %s", err)
		}

		d.SetId(bucket)
	}

	return nil
}

func resourceAwsS3BucketCorsConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	rawCors := d.Get("cors_rule").([]interface{})

	if len(rawCors) == 0 {
		log.Printf("[DEBUG] S3 bucket: %s, delete CORS", bucket)

		_, err := retryOnAwsCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
			return s3conn.DeleteBucketCors(&s3.DeleteBucketCorsInput{
				Bucket: aws.String(bucket),
			})
		})
		if err != nil {
			return fmt.Errorf("Error deleting S3 CORS: %s", err)
		}
	}

	return nil
}
