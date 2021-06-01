package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/glue/waiter"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsGlueDevEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueDevEndpointCreate,
		Read:   resourceAwsGlueDevEndpointRead,
		Update: resourceAwsDevEndpointUpdate,
		Delete: resourceAwsDevEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"extra_jars_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"extra_python_libs_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"glue_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\w+\.\w+$`), "must match version pattern X.X"),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"number_of_nodes": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"number_of_workers", "worker_type"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return new == "0"
				},
				ValidateFunc: validation.IntAtLeast(2),
			},
			"number_of_workers": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntAtLeast(2),
				ConflictsWith: []string{"number_of_nodes"},
			},
			"public_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"public_keys"},
			},
			"public_keys": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
				ConflictsWith: []string{"public_key"},
				MaxItems:      5,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Set:          schema.HashString,
				RequiredWith: []string{"subnet_id"},
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"security_group_ids"},
			},
			"tags": tagsSchema(),
			"private_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"yarn_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zeppelin_remote_spark_interpreter_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"worker_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(glue.WorkerType_Values(), false),
				ConflictsWith: []string{"number_of_nodes"},
				ForceNew:      true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGlueDevEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.CreateDevEndpointInput{
		EndpointName: aws.String(name),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
	}

	if v, ok := d.GetOk("arguments"); ok {
		input.Arguments = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("extra_jars_s3_path"); ok {
		input.ExtraJarsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("extra_python_libs_s3_path"); ok {
		input.ExtraPythonLibsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("number_of_nodes"); ok {
		input.NumberOfNodes = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("public_key"); ok {
		input.PublicKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("public_keys"); ok {
		publicKeys := expandStringSet(v.(*schema.Set))
		input.PublicKeys = publicKeys
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		securityGroupIDs := expandStringSet(v.(*schema.Set))
		input.SecurityGroupIds = securityGroupIDs
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Dev Endpoint: %#v", *input)
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.CreateDevEndpoint(input)
		if err != nil {
			// Retry for IAM eventual consistency
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "should be given assume role permissions for Glue Service") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "is not authorized to perform") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "S3 endpoint and NAT validation has failed for subnetId") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateDevEndpoint(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Glue Dev Endpoint: %s", err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] Waiting for Glue Dev Endpoint (%s) to become available", d.Id())
	if _, err := waiter.GlueDevEndpointCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error while waiting for Glue Dev Endpoint (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsGlueDevEndpointRead(d, meta)
}

func resourceAwsGlueDevEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &glue.GetDevEndpointInput{
		EndpointName: aws.String(d.Id()),
	}

	output, err := conn.GetDevEndpoint(request)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Dev Endpoint (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	endpoint := output.DevEndpoint
	if endpoint == nil {
		log.Printf("[WARN] Glue Dev Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	endpointARN := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("devEndpoint/%s", d.Id()),
	}.String()

	if err := d.Set("arn", endpointARN); err != nil {
		return fmt.Errorf("error setting arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("arguments", aws.StringValueMap(endpoint.Arguments)); err != nil {
		return fmt.Errorf("error setting arguments for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("availability_zone", endpoint.AvailabilityZone); err != nil {
		return fmt.Errorf("error setting availability_zone for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("extra_jars_s3_path", endpoint.ExtraJarsS3Path); err != nil {
		return fmt.Errorf("error setting extra_jars_s3_path for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("extra_python_libs_s3_path", endpoint.ExtraPythonLibsS3Path); err != nil {
		return fmt.Errorf("error setting extra_python_libs_s3_path for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("failure_reason", endpoint.FailureReason); err != nil {
		return fmt.Errorf("error setting failure_reason for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("glue_version", endpoint.GlueVersion); err != nil {
		return fmt.Errorf("error setting glue_version for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("name", endpoint.EndpointName); err != nil {
		return fmt.Errorf("error setting name for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("number_of_nodes", endpoint.NumberOfNodes); err != nil {
		return fmt.Errorf("error setting number_of_nodes for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("number_of_workers", endpoint.NumberOfWorkers); err != nil {
		return fmt.Errorf("error setting number_of_workers for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("private_address", endpoint.PrivateAddress); err != nil {
		return fmt.Errorf("error setting private_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_address", endpoint.PublicAddress); err != nil {
		return fmt.Errorf("error setting public_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_key", endpoint.PublicKey); err != nil {
		return fmt.Errorf("error setting public_key for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("public_keys", flattenStringSet(endpoint.PublicKeys)); err != nil {
		return fmt.Errorf("error setting public_keys for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("role_arn", endpoint.RoleArn); err != nil {
		return fmt.Errorf("error setting role_arn for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_configuration", endpoint.SecurityConfiguration); err != nil {
		return fmt.Errorf("error setting security_configuration for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_group_ids", flattenStringSet(endpoint.SecurityGroupIds)); err != nil {
		return fmt.Errorf("error setting security_group_ids for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("status", endpoint.Status); err != nil {
		return fmt.Errorf("error setting status for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("subnet_id", endpoint.SubnetId); err != nil {
		return fmt.Errorf("error setting subnet_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("vpc_id", endpoint.VpcId); err != nil {
		return fmt.Errorf("error setting vpc_id for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("worker_type", endpoint.WorkerType); err != nil {
		return fmt.Errorf("error setting worker_type for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	tags, err := keyvaluetags.GlueListTags(conn, endpointARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Dev Endpoint (%s): %s", endpointARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("yarn_endpoint_address", endpoint.YarnEndpointAddress); err != nil {
		return fmt.Errorf("error setting yarn_endpoint_address for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	if err := d.Set("zeppelin_remote_spark_interpreter_port", endpoint.ZeppelinRemoteSparkInterpreterPort); err != nil {
		return fmt.Errorf("error setting zeppelin_remote_spark_interpreter_port for Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsDevEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.UpdateDevEndpointInput{
		EndpointName: aws.String(d.Get("name").(string)),
	}

	hasChanged := false

	customLibs := &glue.DevEndpointCustomLibraries{}

	if d.HasChange("arguments") {
		oldRaw, newRaw := d.GetChange("arguments")
		old := oldRaw.(map[string]interface{})
		new := newRaw.(map[string]interface{})
		create, remove, _ := diffStringMaps(old, new)

		removeKeys := make([]*string, 0)
		for k := range remove {
			removeKeys = append(removeKeys, &k)
		}

		input.AddArguments = create
		input.DeleteArguments = removeKeys

		hasChanged = true
	}

	if d.HasChange("extra_jars_s3_path") {
		customLibs.ExtraJarsS3Path = aws.String(d.Get("extra_jars_s3_path").(string))
		input.CustomLibraries = customLibs
		input.UpdateEtlLibraries = aws.Bool(true)

		hasChanged = true
	}

	if d.HasChange("extra_python_libs_s3_path") {
		customLibs.ExtraPythonLibsS3Path = aws.String(d.Get("extra_python_libs_s3_path").(string))
		input.CustomLibraries = customLibs
		input.UpdateEtlLibraries = aws.Bool(true)

		hasChanged = true
	}

	if d.HasChange("public_key") {
		input.PublicKey = aws.String(d.Get("public_key").(string))

		hasChanged = true
	}

	if d.HasChange("public_keys") {
		o, n := d.GetChange("public_keys")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := os.Difference(ns)
		create := ns.Difference(os)

		input.AddPublicKeys = expandStringSet(create)
		log.Printf("[DEBUG] expectedCreate public keys: %v", create)

		input.DeletePublicKeys = expandStringSet(remove)
		log.Printf("[DEBUG] remove public keys: %v", remove)

		hasChanged = true
	}

	if hasChanged {
		log.Printf("[DEBUG] Updating Glue Dev Endpoint: %s", input)
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			_, err := conn.UpdateDevEndpoint(input)
			if err != nil {
				if isAWSErr(err, glue.ErrCodeInvalidInputException, "another concurrent update operation") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}
			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.UpdateDevEndpoint(input)
		}

		if err != nil {
			return fmt.Errorf("error updating Glue Dev Endpoint: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGlueDevEndpointRead(d, meta)
}

func resourceAwsDevEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	deleteOpts := &glue.DeleteDevEndpointInput{
		EndpointName: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting Glue Dev Endpoint: %s", d.Id())

	_, err := conn.DeleteDevEndpoint(deleteOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}

		return fmt.Errorf("error deleting Glue Dev Endpoint (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for Glue Dev Endpoint (%s) to become terminated", d.Id())
	if _, err := waiter.GlueDevEndpointDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}

		return fmt.Errorf("error while waiting for Glue Dev Endpoint (%s) to become terminated: %s", d.Id(), err)
	}

	return nil
}
