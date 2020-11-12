package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsNetworkManagerGlobalNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkManagerGlobalNetworkCreate,
		Read:   resourceAwsNetworkManagerGlobalNetworkRead,
		Update: resourceAwsNetworkManagerGlobalNetworkUpdate,
		Delete: resourceAwsNetworkManagerGlobalNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsNetworkManagerGlobalNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.CreateGlobalNetworkInput{
		Description: aws.String(d.Get("description").(string)),
		Tags:        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NetworkmanagerTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Global Network: %s", input)
	var output *networkmanager.CreateGlobalNetworkOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.CreateGlobalNetwork(input)
		if err != nil {
			if isAWSErr(err, "ValidationException", "Resource already exists with ID") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = conn.CreateGlobalNetwork(input)
	}
	if err != nil {
		return fmt.Errorf("error creating Network Manager Global Network: %s", err)
	}

	d.SetId(aws.StringValue(output.GlobalNetwork.GlobalNetworkId))

	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.GlobalNetworkStatePending},
		Target:  []string{networkmanager.GlobalNetworkStateAvailable},
		Refresh: networkmanagerGlobalNetworkRefreshFunc(conn, aws.StringValue(output.GlobalNetwork.GlobalNetworkId)),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for networkmanager Global Network (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsNetworkManagerGlobalNetworkRead(d, meta)
}

func resourceAwsNetworkManagerGlobalNetworkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	globalNetwork, err := networkmanagerDescribeGlobalNetwork(conn, d.Id())

	if isAWSErr(err, "InvalidGlobalNetworkID.NotFound", "") {
		log.Printf("[WARN] networkmanager Global Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading networkmanager Global Network: %s", err)
	}

	if globalNetwork == nil {
		log.Printf("[WARN] networkmanager Global Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(globalNetwork.State) == networkmanager.GlobalNetworkStateDeleting {
		log.Printf("[WARN] networkmanager Global Network (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(globalNetwork.State))
		d.SetId("")
		return nil
	}

	d.Set("arn", globalNetwork.GlobalNetworkArn)
	d.Set("description", globalNetwork.Description)

	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(globalNetwork.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsNetworkManagerGlobalNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	if d.HasChange("description") {
		request := &networkmanager.UpdateGlobalNetworkInput{
			Description:     aws.String(d.Get("description").(string)),
			GlobalNetworkId: aws.String(d.Id()),
		}

		_, err := conn.UpdateGlobalNetwork(request)
		if err != nil {
			return fmt.Errorf("Failure updating Network Manager Global Network description: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.NetworkmanagerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating networkmanager Global Network (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsNetworkManagerGlobalNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.DeleteGlobalNetworkInput{
		GlobalNetworkId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting networkmanager Global Network (%s): %s", d.Id(), input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteGlobalNetwork(input)

		if isAWSErr(err, "IncorrectState", "has non-deleted Transit Gateway Registrations") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Customer Gateway Associations") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Device") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Link") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Site") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteGlobalNetwork(input)
	}

	if isAWSErr(err, "InvalidGlobalNetworkID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting networkmanager Global Network: %s", err)
	}

	if err := waitForNetworkManagerGlobalNetworkDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for networkmanager Global Network (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func networkmanagerGlobalNetworkRefreshFunc(conn *networkmanager.NetworkManager, globalNetworkID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		globalNetwork, err := networkmanagerDescribeGlobalNetwork(conn, globalNetworkID)

		if isAWSErr(err, "InvalidGlobalNetworkID.NotFound", "") {
			return nil, "DELETED", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading NetworkManager Global Network (%s): %s", globalNetworkID, err)
		}

		if globalNetwork == nil {
			return nil, "DELETED", nil
		}

		return globalNetwork, aws.StringValue(globalNetwork.State), nil
	}
}

func networkmanagerDescribeGlobalNetwork(conn *networkmanager.NetworkManager, globalNetworkID string) (*networkmanager.GlobalNetwork, error) {
	input := &networkmanager.DescribeGlobalNetworksInput{
		GlobalNetworkIds: []*string{aws.String(globalNetworkID)},
	}

	log.Printf("[DEBUG] Reading NetworkManager Global Network (%s): %s", globalNetworkID, input)
	for {
		output, err := conn.DescribeGlobalNetworks(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.GlobalNetworks) == 0 {
			return nil, nil
		}

		for _, globalNetwork := range output.GlobalNetworks {
			if globalNetwork == nil {
				continue
			}

			if aws.StringValue(globalNetwork.GlobalNetworkId) == globalNetworkID {
				return globalNetwork, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func waitForNetworkManagerGlobalNetworkDeletion(conn *networkmanager.NetworkManager, globalNetworkID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			networkmanager.GlobalNetworkStateAvailable,
			networkmanager.GlobalNetworkStateDeleting,
		},
		Target:         []string{""},
		Refresh:        networkmanagerGlobalNetworkRefreshFunc(conn, globalNetworkID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for NetworkManager Global Network (%s) deletion", globalNetworkID)
	_, err := stateConf.WaitForState()

	if isResourceNotFoundError(err) {
		return nil
	}

	return err
}
