package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
)

func resourceAwsCognitoUserPoolSchemaCustomAttributes() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolSchemaCustomAttributeAdd,
		Read:   resourceAwsCognitoUserPoolSchemaAttributRead,
		Update: resourceAwsCognitoUserPoolSchemaCustomAttributeAdd,
		Delete: resourceAwsCognitoUserPoolSchemaCustomAttributDelete,

		Schema: map[string]*schema.Schema{
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"schema": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCognitoUserPoolSchemaName,
						},
						"attribute_data_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentityprovider.AttributeDataTypeString,
								cognitoidentityprovider.AttributeDataTypeNumber,
								cognitoidentityprovider.AttributeDataTypeDateTime,
								cognitoidentityprovider.AttributeDataTypeBoolean,
							}, false),
						},
						"developer_only_attribute": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"mutable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"number_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"string_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_length": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_length": {
										Type:     schema.TypeString,
										Optional: true,
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

func resourceAwsCognitoUserPoolSchemaCustomAttributeAdd(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	log.Printf("[DEBUG] Adding attributes.")

	userPoolId := d.Get("user_pool_id").(string)
	params := &cognitoidentityprovider.AddCustomAttributesInput{
		UserPoolId: aws.String(userPoolId),
	}

	if v, ok := d.GetOk("schema"); ok {
		configs := v.(*schema.Set).List()
		params.CustomAttributes = expandCognitoUserPoolSchema(configs)
	}

	attributeMap, err := _getCurrentAttributeMapFromUserPool(d, meta)
	if err != nil {
		return nil
	}

	params.CustomAttributes = _filterCustomAttributes(params.CustomAttributes, attributeMap)

	log.Printf("[DEBUG] Attributes to add: %s", params)
	if len(params.CustomAttributes) == 0 {
		d.SetId(userPoolId)
		return nil
	}

	_, er := conn.AddCustomAttributes(params)

	if er != nil {
		return errwrap.Wrapf("Error creating Cognito User Pool Custom Attribute: {{err}}", er)
	}

	log.Printf("[DEBUG] Attributes Added Successfully...")

	return resourceAwsCognitoUserPoolSchemaAttributRead(d, meta)
}

func resourceAwsCognitoUserPoolSchemaAttributRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	userPoolId := d.Get("user_pool_id").(string)
	params := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(userPoolId),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool: %s", params)

	resp, err := conn.DescribeUserPool(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Cognito User Pool %s is already gone", userPoolId)
			d.SetId(userPoolId)
			return nil
		}
		return err
	}
	d.SetId(*resp.UserPool.Id)
	return nil
}

func resourceAwsCognitoUserPoolSchemaCustomAttributDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("delete custom attribute operation is not supported")
	return nil
}

func _getCurrentAttributeMapFromUserPool(d *schema.ResourceData, meta interface{}) (map[string]struct{}, error) {
	conn := meta.(*AWSClient).cognitoidpconn

	userPoolId := d.Get("user_pool_id").(string)

	params1 := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(userPoolId),
	}

	resp, err := conn.DescribeUserPool(params1)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Cognito User Pool %s is not found", userPoolId)
		}
		d.SetId(userPoolId)
		return nil, err
	}

	attributeMap := make(map[string]struct{}, len(resp.UserPool.SchemaAttributes))
	for _, sa := range resp.UserPool.SchemaAttributes {
		attributeMap[*sa.Name] = struct{}{}
	}

	return attributeMap, nil
}

// if attribute already exist in user pool remove it from the request attributes.
func _filterCustomAttributes(attr []*cognitoidentityprovider.SchemaAttributeType, attributeMap map[string]struct{}) []*cognitoidentityprovider.SchemaAttributeType {
	i := 0
	for _, p := range attr {
		_, s := attributeMap[*p.Name]
		if !s {
			_, c := attributeMap["custom:"+*p.Name]
			if !c {
				attr[i] = p
				i++
			}
		}
	}

	attr = attr[:i]
	return attr
}
