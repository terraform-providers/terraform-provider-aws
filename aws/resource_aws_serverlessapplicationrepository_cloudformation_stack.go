package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	serverlessrepository "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	cffinder "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/finder"
	cfwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/serverlessapplicationrepository/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/serverlessapplicationrepository/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	serverlessApplicationRepositoryCloudFormationStackNamePrefix = "serverlessrepo-"

	serverlessApplicationRepositoryCloudFormationStackTagApplicationID   = "serverlessrepo:applicationId"
	serverlessApplicationRepositoryCloudFormationStackTagSemanticVersion = "serverlessrepo:semanticVersion"
)

func resourceAwsServerlessApplicationRepositoryCloudFormationStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServerlessApplicationRepositoryCloudFormationStackCreate,
		Read:   resourceAwsServerlessApplicationRepositoryCloudFormationStackRead,
		Update: resourceAwsServerlessApplicationRepositoryCloudFormationStackUpdate,
		Delete: resourceAwsServerlessApplicationRepositoryCloudFormationStackDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsServerlessApplicationRepositoryCloudFormationStackImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.CloudFormationStackCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(waiter.CloudFormationStackUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(waiter.CloudFormationStackDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(serverlessrepository.Capability_Values(), false),
				},
				Set: schema.HashString,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"semantic_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsServerlessApplicationRepositoryCloudFormationStackCreate(d *schema.ResourceData, meta interface{}) error {
	cfConn := meta.(*AWSClient).cfconn

	changeSet, err := createServerlessApplicationRepositoryCloudFormationChangeSet(d, meta.(*AWSClient))
	if err != nil {
		return fmt.Errorf("error creating Serverless Application Repository CloudFormation change set: %w", err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) change set created", d.Id())

	d.SetId(aws.StringValue(changeSet.StackId))

	requestToken := resource.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSet.ChangeSetId,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository CloudFormation change set: %s", executeRequest)
	_, err = cfConn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("executing Serverless Application Repository CloudFormation Stack (%s) change set failed: %w", d.Id(), err)
	}

	_, err = cfwaiter.StackCreated(cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack (%s) creation: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) created", d.Id())

	return resourceAwsServerlessApplicationRepositoryCloudFormationStackRead(d, meta)
}

func resourceAwsServerlessApplicationRepositoryCloudFormationStackRead(d *schema.ResourceData, meta interface{}) error {
	serverlessConn := meta.(*AWSClient).serverlessapplicationrepositoryconn
	cfConn := meta.(*AWSClient).cfconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	stack, err := cffinder.Stack(cfConn, d.Id())
	if tfresource.NotFound(err) {
		log.Printf("[WARN] Serverless Application Repository CloudFormation Stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): %w", d.Id(), err)
	}

	// Serverless Application Repo prefixes the stack name with "serverlessrepo-", so remove it from the saved string
	stackName := strings.TrimPrefix(aws.StringValue(stack.StackName), serverlessApplicationRepositoryCloudFormationStackNamePrefix)
	d.Set("name", &stackName)

	tags := keyvaluetags.CloudformationKeyValueTags(stack.Tags)
	var applicationID, semanticVersion string
	if v, ok := tags[serverlessApplicationRepositoryCloudFormationStackTagApplicationID]; ok {
		applicationID = aws.StringValue(v.Value)
		d.Set("application_id", applicationID)
	} else {
		return fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): missing required tag \"%s\"", d.Id(), serverlessApplicationRepositoryCloudFormationStackTagApplicationID)
	}
	if v, ok := tags[serverlessApplicationRepositoryCloudFormationStackTagSemanticVersion]; ok {
		semanticVersion = aws.StringValue(v.Value)
		d.Set("semantic_version", semanticVersion)
	} else {
		return fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): missing required tag \"%s\"", d.Id(), serverlessApplicationRepositoryCloudFormationStackTagSemanticVersion)
	}

	if err = d.Set("tags", tags.IgnoreServerlessApplicationRepository().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("failed to set tags: %w", err)
	}

	if err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs)); err != nil {
		return fmt.Errorf("failed to set outputs: %w", err)
	}

	getApplicationOutput, err := finder.Application(serverlessConn, applicationID, semanticVersion)
	if err != nil {
		return fmt.Errorf("error getting Serverless Application Repository application (%s, v%s): %w", applicationID, semanticVersion, err)
	}

	if getApplicationOutput == nil || getApplicationOutput.Version == nil {
		return fmt.Errorf("error getting Serverless Application Repository application (%s, v%s): empty response", applicationID, semanticVersion)
	}

	version := getApplicationOutput.Version

	if err = d.Set("parameters", flattenNonDefaultServerlessApplicationCloudFormationParameters(stack.Parameters, version.ParameterDefinitions)); err != nil {
		return fmt.Errorf("failed to set parameters: %w", err)
	}

	if err = d.Set("capabilities", flattenServerlessRepositoryStackCapabilities(stack.Capabilities, version.RequiredCapabilities)); err != nil {
		return fmt.Errorf("failed to set capabilities: %w", err)
	}

	return nil
}

func flattenNonDefaultServerlessApplicationCloudFormationParameters(cfParams []*cloudformation.Parameter, rawParameterDefinitions []*serverlessrepository.ParameterDefinition) map[string]interface{} {
	parameterDefinitions := flattenServerlessRepositoryParameterDefinitions(rawParameterDefinitions)
	params := make(map[string]interface{}, len(cfParams))
	for _, p := range cfParams {
		key := aws.StringValue(p.ParameterKey)
		value := aws.StringValue(p.ParameterValue)
		if value != aws.StringValue(parameterDefinitions[key].DefaultValue) {
			params[key] = value
		}
	}
	return params
}

func flattenServerlessRepositoryParameterDefinitions(parameterDefinitions []*serverlessrepository.ParameterDefinition) map[string]*serverlessrepository.ParameterDefinition {
	result := make(map[string]*serverlessrepository.ParameterDefinition, len(parameterDefinitions))
	for _, p := range parameterDefinitions {
		result[aws.StringValue(p.Name)] = p
	}
	return result
}

func resourceAwsServerlessApplicationRepositoryCloudFormationStackUpdate(d *schema.ResourceData, meta interface{}) error {
	cfConn := meta.(*AWSClient).cfconn

	changeSet, err := createServerlessApplicationRepositoryCloudFormationChangeSet(d, meta.(*AWSClient))
	if err != nil {
		return fmt.Errorf("error creating Serverless Application Repository CloudFormation Stack (%s) change set: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) change set created", d.Id())

	requestToken := resource.UniqueId()
	executeRequest := cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      changeSet.ChangeSetId,
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Executing Serverless Application Repository CloudFormation change set: %s", executeRequest)
	_, err = cfConn.ExecuteChangeSet(&executeRequest)
	if err != nil {
		return fmt.Errorf("executing Serverless Application Repository CloudFormation change set failed: %w", err)
	}

	_, err = cfwaiter.StackUpdated(cfConn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack (%s) update: %w", d.Id(), err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation Stack (%s) updated", d.Id())

	return resourceAwsServerlessApplicationRepositoryCloudFormationStackRead(d, meta)
}

func resourceAwsServerlessApplicationRepositoryCloudFormationStackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	requestToken := resource.UniqueId()
	input := &cloudformation.DeleteStackInput{
		StackName:          aws.String(d.Id()),
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Deleting Serverless Application Repository CloudFormation stack %s", input)
	_, err := conn.DeleteStack(input)
	if tfawserr.ErrCodeEquals(err, "ValidationError") {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = cfwaiter.StackDeleted(conn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("error waiting for Serverless Application Repository CloudFormation Stack deletion: %w", err)
	}

	log.Printf("[INFO] Serverless Application Repository CloudFormation stack (%s) deleted", d.Id())

	return nil
}

func resourceAwsServerlessApplicationRepositoryCloudFormationStackImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	stackID := d.Id()

	// If this isn't an ARN, it's the stack name
	if _, err := arn.Parse(stackID); err != nil {
		if !strings.HasPrefix(stackID, serverlessApplicationRepositoryCloudFormationStackNamePrefix) {
			stackID = serverlessApplicationRepositoryCloudFormationStackNamePrefix + stackID
		}
	}

	cfConn := meta.(*AWSClient).cfconn
	stack, err := cffinder.Stack(cfConn, stackID)
	if err != nil {
		return nil, fmt.Errorf("error describing Serverless Application Repository CloudFormation Stack (%s): %w", stackID, err)
	}

	d.SetId(aws.StringValue(stack.StackId))

	return []*schema.ResourceData{d}, nil
}

func createServerlessApplicationRepositoryCloudFormationChangeSet(d *schema.ResourceData, client *AWSClient) (*cloudformation.DescribeChangeSetOutput, error) {
	serverlessConn := client.serverlessapplicationrepositoryconn
	cfConn := client.cfconn

	stackName := d.Get("name").(string)
	changeSetRequest := serverlessrepository.CreateCloudFormationChangeSetRequest{
		StackName:     aws.String(stackName),
		ApplicationId: aws.String(d.Get("application_id").(string)),
		Capabilities:  expandStringSet(d.Get("capabilities").(*schema.Set)),
		Tags:          keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreServerlessApplicationRepository().ServerlessapplicationrepositoryTags(),
	}
	if v, ok := d.GetOk("semantic_version"); ok {
		changeSetRequest.SemanticVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("parameters"); ok {
		changeSetRequest.ParameterOverrides = expandServerlessRepositoryCloudFormationChangeSetParameters(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Serverless Application Repository CloudFormation change set: %s", changeSetRequest)
	changeSetResponse, err := serverlessConn.CreateCloudFormationChangeSet(&changeSetRequest)
	if err != nil {
		return nil, err
	}

	return cfwaiter.ChangeSetCreated(cfConn, aws.StringValue(changeSetResponse.StackId), aws.StringValue(changeSetResponse.ChangeSetId))
}

func expandServerlessRepositoryCloudFormationChangeSetParameters(params map[string]interface{}) []*serverlessrepository.ParameterValue {
	var appParams []*serverlessrepository.ParameterValue
	for k, v := range params {
		appParams = append(appParams, &serverlessrepository.ParameterValue{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return appParams
}

func flattenServerlessRepositoryStackCapabilities(stackCapabilities []*string, applicationRequiredCapabilities []*string) *schema.Set {
	// We need to preserve "CAPABILITY_RESOURCE_POLICY" if it has been set. It is not
	// returned by the CloudFormation APIs.
	capabilities := flattenStringSet(stackCapabilities)
	for _, capability := range applicationRequiredCapabilities {
		if aws.StringValue(capability) == serverlessrepository.CapabilityCapabilityResourcePolicy {
			capabilities.Add(serverlessrepository.CapabilityCapabilityResourcePolicy)
			break
		}
	}
	return capabilities
}
