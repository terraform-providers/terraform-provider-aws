package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tftransfer "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsTransferServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTransferServerCreate,
		Read:   resourceAwsTransferServerRead,
		Update: resourceAwsTransferServerUpdate,
		Delete: resourceAwsTransferServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"domain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      transfer.DomainS3,
				ValidateFunc: validation.StringInSlice(transfer.Domain_Values(), false),
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint_details": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_allocation_ids": {
							Type:          schema.TypeSet,
							Optional:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Set:           schema.HashString,
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
						"subnet_ids": {
							Type:          schema.TypeSet,
							Optional:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							Set:           schema.HashString,
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
						"vpc_endpoint_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"endpoint_details.0.address_allocation_ids", "endpoint_details.0.subnet_ids", "endpoint_details.0.vpc_id"},
							Computed:      true,
						},
						"vpc_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  validation.NoZeroValues,
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
					},
				},
			},

			"endpoint_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      transfer.EndpointTypePublic,
				ValidateFunc: validation.StringInSlice(transfer.EndpointType_Values(), false),
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"host_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 4096),
			},

			"host_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"identity_provider_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      transfer.IdentityProviderTypeServiceManaged,
				ValidateFunc: validation.StringInSlice(transfer.IdentityProviderType_Values(), false),
			},

			"invocation_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"logging_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"protocols": {
				Type:     schema.TypeSet,
				MinItems: 1,
				MaxItems: 3,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(transfer.Protocol_Values(), false),
				},
			},

			"security_policy_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "TransferSecurityPolicy-2018-11",
				ValidateFunc: validation.StringInSlice([]string{
					"TransferSecurityPolicy-2018-11",
					"TransferSecurityPolicy-2020-06",
					"TransferSecurityPolicy-FIPS-2020-06",
				}, false),
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),

			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsTransferServerCreate(d *schema.ResourceData, meta interface{}) error {
	updateAfterCreate := false
	conn := meta.(*AWSClient).transferconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &transfer.CreateServerInput{}

	if v, ok := d.GetOk("certificate"); ok {
		input.Certificate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain"); ok {
		input.Domain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("endpoint_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EndpointDetails = expandTransferEndpointDetails(v.([]interface{})[0].(map[string]interface{}))

		// Prevent the following error: InvalidRequestException: AddressAllocationIds cannot be set in CreateServer
		// Reference: https://docs.aws.amazon.com/transfer/latest/userguide/API_EndpointDetails.html#TransferFamily-Type-EndpointDetails-AddressAllocationIds
		if input.EndpointDetails != nil && len(input.EndpointDetails.AddressAllocationIds) > 0 {
			input.EndpointDetails.AddressAllocationIds = nil
			updateAfterCreate = true
		}
	}

	if v, ok := d.GetOk("endpoint_type"); ok {
		input.EndpointType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("host_key"); ok {
		input.HostKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("identity_provider_type"); ok {
		input.IdentityProviderType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_role"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &transfer.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.InvocationRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_role"); ok {
		input.LoggingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("protocols"); ok && v.(*schema.Set).Len() > 0 {
		input.Protocols = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("security_policy_name"); ok {
		input.SecurityPolicyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("url"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &transfer.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.Url = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().TransferTags()
	}

	log.Printf("[DEBUG] Creating Transfer Server: %s", input)
	output, err := conn.CreateServer(input)

	if err != nil {
		return fmt.Errorf("error creating Transfer Server: %w", err)
	}

	d.SetId(aws.StringValue(output.ServerId))

	_, err = waiter.ServerCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for Transfer Server (%s) to create: %w", d.Id(), err)
	}

	if updateAfterCreate {
		if err := stopTransferServer(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return err
		}

		input := &transfer.UpdateServerInput{
			ServerId:        aws.String(d.Id()),
			EndpointDetails: expandTransferEndpointDetails(d.Get("endpoint_details").([]interface{})[0].(map[string]interface{})),
		}

		if err := updateTransferServer(conn, input); err != nil {
			return err
		}

		if err := startTransferServer(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return err
		}
	}

	return resourceAwsTransferServerRead(d, meta)
}

func resourceAwsTransferServerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := finder.ServerByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Server (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Transfer Server (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("certificate", output.Certificate)
	d.Set("domain", output.Domain)
	d.Set("endpoint", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.server.transfer", d.Id())))
	if output.EndpointDetails != nil {
		if err := d.Set("endpoint_details", []interface{}{flattenTransferEndpointDetails(output.EndpointDetails)}); err != nil {
			return fmt.Errorf("error setting endpoint_details: %w", err)
		}
	} else {
		d.Set("endpoint_details", nil)
	}
	d.Set("endpoint_type", output.EndpointType)
	d.Set("host_key_fingerprint", output.HostKeyFingerprint)
	d.Set("identity_provider_type", output.IdentityProviderType)
	if output.IdentityProviderDetails != nil {
		d.Set("invocation_role", output.IdentityProviderDetails.InvocationRole)
	} else {
		d.Set("invocation_role", "")
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("protocols", aws.StringValueSlice(output.Protocols))
	d.Set("security_policy_name", output.SecurityPolicyName)
	if output.IdentityProviderDetails != nil {
		d.Set("url", output.IdentityProviderDetails.Url)
	} else {
		d.Set("url", "")
	}

	tags := keyvaluetags.TransferKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsTransferServerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	if d.HasChangesExcept("tags", "tags_all") {
		stopFlag := false

		input := &transfer.UpdateServerInput{
			ServerId: aws.String(d.Id()),
		}

		if d.HasChange("logging_role") {
			input.LoggingRole = aws.String(d.Get("logging_role").(string))
		}

		if d.HasChange("security_policy_name") {
			input.SecurityPolicyName = aws.String(d.Get("security_policy_name").(string))
		}

		if d.HasChanges("invocation_role", "url") {
			identityProviderDetails := &transfer.IdentityProviderDetails{}

			if attr, ok := d.GetOk("invocation_role"); ok {
				identityProviderDetails.InvocationRole = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("url"); ok {
				identityProviderDetails.Url = aws.String(attr.(string))
			}

			input.IdentityProviderDetails = identityProviderDetails
		}

		if d.HasChange("endpoint_type") {
			input.EndpointType = aws.String(d.Get("endpoint_type").(string))
		}

		if d.HasChange("certificate") {
			input.Certificate = aws.String(d.Get("certificate").(string))
		}

		if d.HasChange("protocols") {
			input.Protocols = expandStringSet(d.Get("protocols").(*schema.Set))
		}

		if d.HasChange("endpoint_details") {
			if v, ok := d.GetOk("endpoint_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.EndpointDetails = expandTransferEndpointDetails(v.([]interface{})[0].(map[string]interface{}))
			}

			// Prevent the following error: InvalidRequestException: Server must be OFFLINE to change AddressAllocationIds
			if d.HasChange("endpoint_details.0.address_allocation_ids") {
				stopFlag = true
			}
		}

		if d.HasChange("host_key") {
			if attr, ok := d.GetOk("host_key"); ok {
				input.HostKey = aws.String(attr.(string))
			}
		}

		if stopFlag {
			if err := stopTransferServer(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return err
			}
		}

		log.Printf("[DEBUG] Updating Transfer Server: %s", input)
		if err := updateTransferServer(conn, input); err != nil {
			return err
		}

		if stopFlag {
			if err := startTransferServer(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.TransferUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsTransferServerRead(d, meta)
}

func resourceAwsTransferServerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).transferconn

	if d.Get("force_destroy").(bool) {
		input := &transfer.ListUsersInput{
			ServerId: aws.String(d.Id()),
		}
		var deletionErrs *multierror.Error

		err := conn.ListUsersPages(input, func(page *transfer.ListUsersOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, user := range page.Users {
				resourceID := tftransfer.UserCreateResourceID(d.Id(), aws.StringValue(user.UserName))

				r := resourceAwsTransferUser()
				d := r.Data(nil)
				d.SetId(resourceID)
				err := r.Delete(d, meta)

				if err != nil {
					deletionErrs = multierror.Append(deletionErrs, fmt.Errorf("error deleting Transfer User (%s): %w", resourceID, err))
					continue
				}
			}

			return !lastPage
		})

		if err != nil {
			deletionErrs = multierror.Append(deletionErrs, fmt.Errorf("error listing Transfer Users: %w", err))
		}

		err = deletionErrs.ErrorOrNil()

		if err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Transfer Server (%s)", d.Id())
	_, err := conn.DeleteServer(&transfer.DeleteServerInput{
		ServerId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Transfer Server (%s): %w", d.Id(), err)
	}

	_, err = waiter.ServerDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Transfer Server (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandTransferEndpointDetails(tfMap map[string]interface{}) *transfer.EndpointDetails {
	if tfMap == nil {
		return nil
	}

	apiObject := &transfer.EndpointDetails{}

	if v, ok := tfMap["address_allocation_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AddressAllocationIds = expandStringSet(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = expandStringSet(v)
	}

	if v, ok := tfMap["vpc_endpoint_id"].(string); ok && v != "" {
		apiObject.VpcEndpointId = aws.String(v)
	}

	if v, ok := tfMap["vpc_id"].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenTransferEndpointDetails(apiObject *transfer.EndpointDetails) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AddressAllocationIds; v != nil {
		tfMap["address_allocation_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcEndpointId; v != nil {
		tfMap["vpc_endpoint_id"] = aws.StringValue(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	return tfMap
}

func stopTransferServer(conn *transfer.Transfer, serverID string, timeout time.Duration) error {
	input := &transfer.StopServerInput{
		ServerId: aws.String(serverID),
	}

	if _, err := conn.StopServer(input); err != nil {
		return fmt.Errorf("error stopping Transfer Server (%s): %w", serverID, err)
	}

	if _, err := waiter.ServerStopped(conn, serverID, timeout); err != nil {
		return fmt.Errorf("error waiting for Transfer Server (%s) to stop: %w", serverID, err)
	}

	return nil
}

func startTransferServer(conn *transfer.Transfer, serverID string, timeout time.Duration) error {
	input := &transfer.StartServerInput{
		ServerId: aws.String(serverID),
	}

	if _, err := conn.StartServer(input); err != nil {
		return fmt.Errorf("error starting Transfer Server (%s): %w", serverID, err)
	}

	if _, err := waiter.ServerStarted(conn, serverID, timeout); err != nil {
		return fmt.Errorf("error waiting for Transfer Server (%s) to start: %w", serverID, err)
	}

	return nil
}

func updateTransferServer(conn *transfer.Transfer, input *transfer.UpdateServerInput) error {
	// The Transfer API will return a state of ONLINE for a server before the
	// underlying VPC Endpoint is available and attempting to update the server
	// will return an error until that EC2 API process is complete:
	//   ConflictException: VPC Endpoint state is not yet available
	// To prevent accessing the EC2 API directly to check the VPC Endpoint
	// state, which can require confusing IAM permissions and have other
	// eventual consistency consideration, we retry only via the Transfer API.
	err := resource.Retry(Ec2VpcEndpointCreationTimeout, func() *resource.RetryError {
		_, err := conn.UpdateServer(input)

		if tfawserr.ErrMessageContains(err, transfer.ErrCodeConflictException, "VPC Endpoint state is not yet available") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateServer(input)
	}

	if err != nil {
		return fmt.Errorf("error updating Transfer Server (%s): %w", aws.StringValue(input.ServerId), err)
	}

	return nil
}
