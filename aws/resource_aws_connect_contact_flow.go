package aws

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	// ConnectContactFlowCreateTimeout Timeout for connect flow creation
	ConnectContactFlowCreateTimeout = 1 * time.Minute
	// ConnectContactFlowUpdateTimeout Timeout for connect flow update
	ConnectContactFlowUpdateTimeout = 1 * time.Minute
	// ConnectContactFlowDeleteTimeout Timeout for connect flow deletion
	ConnectContactFlowDeleteTimeout = 1 * time.Minute
)

const awsMutexConnectContactFlowKey = `aws_connect_contact_flow`

func resourceAwsConnectContactFlow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsConnectContactFlowCreate,
		ReadContext:   resourceAwsConnectContactFlowRead,
		UpdateContext: resourceAwsConnectContactFlowUpdate,
		DeleteContext: resourceAwsConnectContactFlowDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(d.Id())

				if err != nil {
					return nil, err
				}

				d.Set("instance_id", instanceID)
				d.Set("contact_flow_id", contactFlowID)
				d.SetId(fmt.Sprintf("%s:%s", instanceID, contactFlowID))

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ConnectContactFlowCreateTimeout),
			Update: schema.DefaultTimeout(ConnectContactFlowUpdateTimeout),
			Delete: schema.DefaultTimeout(ConnectContactFlowDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				ConflictsWith:    []string{"filename"},
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"content"},
			},
			"content_hash": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(connect.ContactFlowType_Values(), false),
				Default:      connect.ContactFlowTypeContactFlow,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsConnectContactFlowParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:connectFlowID", id)
	}

	return parts[0], parts[1], nil
}

func resourceAwsConnectContactFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)
	filename, filenameOK := d.GetOk("filename")
	content, contentOK := d.GetOk("content")

	input := &connect.CreateContactFlowInput{
		Name:        aws.String(name),
		InstanceId:  aws.String(instanceID),
		Description: aws.String(d.Get("description").(string)),
		Type:        aws.String(d.Get("type").(string)),
	}

	if filenameOK {
		// Grab an exclusive lock so that we're only reading one contact flow into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		awsMutexKV.Lock(awsMutexConnectContactFlowKey)
		defer awsMutexKV.Unlock(awsMutexConnectContactFlowKey)
		file, err := resourceAwsConnectContactFlowLoadFileContent(filename.(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("Unable to load %q: %w", filename.(string), err))
		}
		input.Content = aws.String(file)
	} else if contentOK {
		input.Content = aws.String(content.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ConnectTags()
	}

	output, err := conn.CreateContactFlowWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("Error creating contact flow '%s': %w", name, err))
	}
	d.Set("arn", output.ContactFlowArn)
	d.Set("contact_flow_id", output.ContactFlowId)
	d.SetId(fmt.Sprintf("%s:%s", instanceID, d.Get("contact_flow_id").(string)))

	return resourceAwsConnectContactFlowRead(ctx, d, meta)
}

func resourceAwsConnectContactFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeContactFlowWithContext(ctx, &connect.DescribeContactFlowInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
	})
	if isAWSErr(err, connect.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Connect flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error getting connect flow '%s': %w", d.Id(), err))
	}

	d.Set("arn", resp.ContactFlow.Arn)
	d.Set("contact_flow_id", resp.ContactFlow.Id)

	d.Set("name", resp.ContactFlow.Name)
	d.Set("description", resp.ContactFlow.Description)
	d.Set("type", resp.ContactFlow.Type)
	d.Set("content", resp.ContactFlow.Content)

	if err := d.Set("tags", keyvaluetags.ConnectKeyValueTags(resp.ContactFlow.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}
	return nil
}

func resourceAwsConnectContactFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	filename, filenameOK := d.GetOk("filename")
	content, contentOK := d.GetOk("content")

	if d.HasChanges("name", "description") {
		updateMetadataInput := &connect.UpdateContactFlowNameInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
			Name:          aws.String(d.Get("name").(string)),
			Description:   aws.String(d.Get("description").(string)),
		}

		_, updateMetadataInputErr := conn.UpdateContactFlowNameWithContext(ctx, updateMetadataInput)

		if updateMetadataInputErr != nil {
			return diag.FromErr(updateMetadataInputErr)
		}
	}

	if d.HasChanges("content", "content_hash") {
		updateContentInput := &connect.UpdateContactFlowContentInput{
			ContactFlowId: aws.String(contactFlowID),
			InstanceId:    aws.String(instanceID),
		}

		if filenameOK {
			// Grab an exclusive lock so that we're only reading one contact flow into
			// memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			awsMutexKV.Lock(awsMutexConnectContactFlowKey)
			defer awsMutexKV.Unlock(awsMutexConnectContactFlowKey)
			file, err := resourceAwsConnectContactFlowLoadFileContent(filename.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("Unable to load %q: %w", filename.(string), err))
			}
			updateContentInput.Content = aws.String(file)
		} else if contentOK {
			updateContentInput.Content = aws.String(content.(string))
		}

		_, updateContentInputErr := conn.UpdateContactFlowContentWithContext(ctx, updateContentInput)

		if updateContentInputErr != nil {
			return diag.FromErr(updateContentInputErr)
		}
	}
	if d.HasChange("tags") {
		oldTags, newTags := d.GetChange("tags")
		if err := keyvaluetags.ConnectUpdateTags(conn, d.Get("arn").(string), oldTags, newTags); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceAwsConnectContactFlowRead(ctx, d, meta)
}

func resourceAwsConnectContactFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID, contactFlowID, err := resourceAwsConnectContactFlowParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.UpdateContactFlowNameInput{
		ContactFlowId: aws.String(contactFlowID),
		InstanceId:    aws.String(instanceID),
		Name:          aws.String(fmt.Sprintf("%s:%s:%d", "zzTrash", d.Get("name").(string), time.Now().Unix())),
		Description:   aws.String("DELETED"),
	}

	_, delerr := conn.UpdateContactFlowNameWithContext(ctx, input)

	if delerr != nil {
		return diag.FromErr(fmt.Errorf("Unable to delete contact flow: %s", delerr))
	}

	return nil
}

func resourceAwsConnectContactFlowLoadFileContent(filename string) (string, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	content := string(fileContent)
	return content, nil
}
