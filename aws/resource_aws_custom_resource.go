package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCustomResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCustomResourceCreate,
		Read:   resourceAwsCustomResourceRead,
		Update: resourceAwsCustomResourceUpdate,
		Delete: resourceAwsCustomResourceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"service_token": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"old_resource_properties": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
			"data": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func resourceAwsCustomResourceCreate(d *schema.ResourceData, meta interface{}) error {
	d.Set("old_resource_properties", d.Get("resource_properties"))
	data, err := invokeLambda("Create", d, meta)
	if err != nil {
		return err
	}
	d.Set("data", data)
	id := strconv.FormatUint(rand.Uint64(), 10)
	d.SetId(id)
	return resourceAwsCustomResourceRead(d, meta)
}

func resourceAwsCustomResourceRead(d *schema.ResourceData, meta interface{}) error {
	//AWS Custom Resource does not support "Read" https://docs.aws.amazon.com/en_pv/AWSCloudFormation/latest/UserGuide/crpg-ref-requesttypes.html
	return nil
}

func resourceAwsCustomResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("resource_properties") {
		oldResourceProperties, _ := d.GetChange("resource_properties")
		d.Set("oldResourceProperties", oldResourceProperties)
		data, err := invokeLambda("Update", d, meta)
		if err != nil {
			return err
		}

		d.Set("data", data)
	}

	return resourceAwsCustomResourceRead(d, meta)
}

func resourceAwsCustomResourceDelete(d *schema.ResourceData, meta interface{}) error {
	data, err := invokeLambda("Delete", d, meta)
	if err != nil {
		return err
	}

	d.Set("data", data)
	d.SetId("")
	return nil
}

func invokeLambda(requestType string, d *schema.ResourceData, meta interface{}) (map[string]string, error) {
	conn := meta.(*AWSClient).lambdaconn
	serviceToken := d.Get("service_token").(string)
	resourceType := d.Get("resource_type").(string)

	var oldResourceProperties map[string]string
	if v, ok := d.GetOk("old_resource_properties"); ok {
		oldResourceProperties = readProperties(v.(map[string]interface{}))
	}

	var resourceProperties map[string]string
	if v, ok := d.GetOk("resource_properties"); ok {
		resourceProperties = readProperties(v.(map[string]interface{}))
	}

	customResourceRequest := &CustomResourceRequest{
		RequestType:           requestType,
		ResourceType:          resourceType,
		OldResourceProperties: oldResourceProperties,
		ResourceProperties:    resourceProperties,
	}

	payload, _ := json.Marshal(customResourceRequest)

	prettyPayload := string(payload)
	log.Printf("[DEBUG] Input payload to lambda: %s", prettyPayload)

	log.Printf("[DEBUG] %s: Lambda-backed Custom Resource with lambda arn %s", requestType, serviceToken)

	logType := "Tail"
	invokeRequest := &lambda.InvokeInput{
		FunctionName: aws.String(serviceToken),
		Payload:      payload,
		LogType:      &logType,
	}

	var invokeResponse *lambda.InvokeOutput
	invokeResponse, err := conn.Invoke(invokeRequest)

	if err != nil {
		return nil, fmt.Errorf("Error invoking lambda function %s: %s", serviceToken, err)
	}
	if *invokeResponse.StatusCode != 200 {
		return nil, fmt.Errorf("Lambda returned %d status code with error message: %s", *invokeResponse.StatusCode, *invokeResponse.LogResult)
	}
	var customResourceResponse CustomResourceResponse
	err = json.Unmarshal(invokeResponse.Payload, &customResourceResponse)
	if err != nil {
		return nil, fmt.Errorf("Response cannot be unmarshalled into JSON: %v", err)
	}
	log.Printf("[DEBUG] Output from lambda function: %v", customResourceResponse)
	if customResourceResponse.Status == "FAILED" {
		return nil, fmt.Errorf(`Custom resource returned "FAILED" Status code with Reason: %s`, customResourceResponse.Reason)
	}
	data := readProperties(customResourceResponse.Data)
	return data, nil
}

//CustomResourceRequest is an adapter model for requests
type CustomResourceRequest struct {
	RequestType           string
	ResourceType          string
	OldResourceProperties map[string]string
	ResourceProperties    map[string]string
}

//CustomResourceResponse is an adapter model for responses
type CustomResourceResponse struct {
	Status string
	Reason string
	Data   map[string]interface{}
}

func readProperties(ev map[string]interface{}) map[string]string {
	variables := make(map[string]string)
	for k, v := range ev {
		variables[k] = v.(string)
	}
	return variables
}
