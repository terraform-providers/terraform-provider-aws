package aws

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsVpcEndpointService() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointServiceCreate,
		Read:   resourceAwsVpcEndpointServiceRead,
		Update: resourceAwsVpcEndpointServiceUpdate,
		Delete: resourceAwsVpcEndpointServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"acceptance_required": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"allowed_principals": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"base_endpoint_dns_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"manages_vpc_endpoints": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"network_load_balancer_arns": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
				Set: schema.HashString,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsVpcEndpointServiceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.CreateVpcEndpointServiceConfigurationInput{
		AcceptanceRequired:      aws.Bool(d.Get("acceptance_required").(bool)),
		NetworkLoadBalancerArns: expandStringSet(d.Get("network_load_balancer_arns").(*schema.Set)),
		TagSpecifications:       ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), "vpc-endpoint-service"),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Service configuration: %#v", req)
	resp, err := conn.CreateVpcEndpointServiceConfiguration(req)
	if err != nil {
		return fmt.Errorf("Error creating VPC Endpoint Service configuration: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.ServiceConfiguration.ServiceId))

	if err := vpcEndpointServiceWaitUntilAvailable(d, conn); err != nil {
		return err
	}

	if v, ok := d.GetOk("allowed_principals"); ok && v.(*schema.Set).Len() > 0 {
		modifyPermReq := &ec2.ModifyVpcEndpointServicePermissionsInput{
			ServiceId:            aws.String(d.Id()),
			AddAllowedPrincipals: expandStringSet(v.(*schema.Set)),
		}
		log.Printf("[DEBUG] Adding VPC Endpoint Service permissions: %#v", modifyPermReq)
		if _, err := conn.ModifyVpcEndpointServicePermissions(modifyPermReq); err != nil {
			return fmt.Errorf("error adding VPC Endpoint Service permissions: %s", err.Error())
		}
	}

	return resourceAwsVpcEndpointServiceRead(d, meta)
}

func resourceAwsVpcEndpointServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	svcCfgRaw, state, err := vpcEndpointServiceStateRefresh(conn, d.Id())()
	if err != nil && state != ec2.ServiceStateFailed {
		return fmt.Errorf("error reading VPC Endpoint Service (%s): %s", d.Id(), err.Error())
	}

	terminalStates := map[string]bool{
		ec2.ServiceStateDeleted:  true,
		ec2.ServiceStateDeleting: true,
		ec2.ServiceStateFailed:   true,
	}
	if _, ok := terminalStates[state]; ok {
		log.Printf("[WARN] VPC Endpoint Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	svcCfg := svcCfgRaw.(*ec2.ServiceConfiguration)
	d.Set("acceptance_required", svcCfg.AcceptanceRequired)
	err = d.Set("network_load_balancer_arns", flattenStringSet(svcCfg.NetworkLoadBalancerArns))
	if err != nil {
		return fmt.Errorf("error setting network_load_balancer_arns: %s", err)
	}
	err = d.Set("availability_zones", flattenStringSet(svcCfg.AvailabilityZones))
	if err != nil {
		return fmt.Errorf("error setting availability_zones: %s", err)
	}
	err = d.Set("base_endpoint_dns_names", flattenStringSet(svcCfg.BaseEndpointDnsNames))
	if err != nil {
		return fmt.Errorf("error setting base_endpoint_dns_names: %s", err)
	}
	d.Set("manages_vpc_endpoints", svcCfg.ManagesVpcEndpoints)
	d.Set("private_dns_name", svcCfg.PrivateDnsName)
	d.Set("service_name", svcCfg.ServiceName)
	d.Set("service_type", svcCfg.ServiceType[0].ServiceType)
	d.Set("state", svcCfg.ServiceState)
	err = d.Set("tags", keyvaluetags.Ec2KeyValueTags(svcCfg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	if err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	resp, err := conn.DescribeVpcEndpointServicePermissions(&ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Service permissions (%s): %s", d.Id(), err.Error())
	}

	err = d.Set("allowed_principals", flattenVpcEndpointServiceAllowedPrincipals(resp.AllowedPrincipals))
	if err != nil {
		return fmt.Errorf("error setting allowed_principals: %s", err)
	}

	return nil
}

func resourceAwsVpcEndpointServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	svcId := d.Id()

	modifyCfgReq := &ec2.ModifyVpcEndpointServiceConfigurationInput{
		ServiceId: aws.String(svcId),
	}
	modifyCfg := false
	if d.HasChange("acceptance_required") {
		modifyCfgReq.AcceptanceRequired = aws.Bool(d.Get("acceptance_required").(bool))
		modifyCfg = true
	}
	if setVpcEndpointServiceUpdateLists(d, "network_load_balancer_arns",
		&modifyCfgReq.AddNetworkLoadBalancerArns, &modifyCfgReq.RemoveNetworkLoadBalancerArns) {
		modifyCfg = true
	}
	if modifyCfg {
		log.Printf("[DEBUG] Modifying VPC Endpoint Service configuration: %#v", modifyCfgReq)
		if _, err := conn.ModifyVpcEndpointServiceConfiguration(modifyCfgReq); err != nil {
			return fmt.Errorf("Error modifying VPC Endpoint Service configuration: %s", err.Error())
		}
		if err := vpcEndpointServiceWaitUntilAvailable(d, conn); err != nil {
			return err
		}
	}

	modifyPermReq := &ec2.ModifyVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(svcId),
	}
	if setVpcEndpointServiceUpdateLists(d, "allowed_principals",
		&modifyPermReq.AddAllowedPrincipals, &modifyPermReq.RemoveAllowedPrincipals) {
		log.Printf("[DEBUG] Modifying VPC Endpoint Service permissions: %#v", modifyPermReq)
		if _, err := conn.ModifyVpcEndpointServicePermissions(modifyPermReq); err != nil {
			return fmt.Errorf("Error modifying VPC Endpoint Service permissions: %s", err.Error())
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPC Endpoint Service (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsVpcEndpointServiceRead(d, meta)
}

func resourceAwsVpcEndpointServiceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Deleting VPC Endpoint Service: %s", d.Id())
	_, err := conn.DeleteVpcEndpointServiceConfigurations(&ec2.DeleteVpcEndpointServiceConfigurationsInput{
		ServiceIds: aws.StringSlice([]string{d.Id()}),
	})
	if err != nil {
		if isAWSErr(err, "InvalidVpcEndpointServiceId.NotFound", "") {
			log.Printf("[DEBUG] VPC Endpoint Service %s is already gone", d.Id())
		} else {
			return fmt.Errorf("Error deleting VPC Endpoint Service: %s", err.Error())
		}
	}

	if err := waitForVpcEndpointServiceDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for VPC Endpoint Service %s to delete: %s", d.Id(), err.Error())
	}

	return nil
}

func vpcEndpointServiceStateRefresh(conn *ec2.EC2, svcId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Reading VPC Endpoint Service Configuration: %s", svcId)
		resp, err := conn.DescribeVpcEndpointServiceConfigurations(&ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: aws.StringSlice([]string{svcId}),
		})
		if err != nil {
			if isAWSErr(err, "InvalidVpcEndpointServiceId.NotFound", "") {
				return false, ec2.ServiceStateDeleted, nil
			}

			return nil, "", err
		}

		svcCfg := resp.ServiceConfigurations[0]
		state := aws.StringValue(svcCfg.ServiceState)
		// No use in retrying if the endpoint service is in a failed state.
		if state == ec2.ServiceStateFailed {
			return nil, state, errors.New("VPC Endpoint Service is in a failed state")
		}
		return svcCfg, state, nil
	}
}

func vpcEndpointServiceWaitUntilAvailable(d *schema.ResourceData, conn *ec2.EC2) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.ServiceStatePending},
		Target:     []string{ec2.ServiceStateAvailable},
		Refresh:    vpcEndpointServiceStateRefresh(conn, d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for VPC Endpoint Service %s to become available: %s", d.Id(), err.Error())
	}

	return nil
}

func waitForVpcEndpointServiceDeletion(conn *ec2.EC2, serviceID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.ServiceStateAvailable, ec2.ServiceStateDeleting},
		Target:     []string{ec2.ServiceStateDeleted},
		Refresh:    vpcEndpointServiceStateRefresh(conn, serviceID),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func setVpcEndpointServiceUpdateLists(d *schema.ResourceData, key string, a, r *[]*string) bool {
	if !d.HasChange(key) {
		return false
	}

	o, n := d.GetChange(key)
	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	add := expandStringList(ns.Difference(os).List())
	if len(add) > 0 {
		*a = add
	}

	remove := expandStringList(os.Difference(ns).List())
	if len(remove) > 0 {
		*r = remove
	}

	return true
}

func flattenVpcEndpointServiceAllowedPrincipals(allowedPrincipals []*ec2.AllowedPrincipal) *schema.Set {
	vPrincipals := []interface{}{}

	for _, allowedPrincipal := range allowedPrincipals {
		if allowedPrincipal.Principal != nil {
			vPrincipals = append(vPrincipals, aws.StringValue(allowedPrincipal.Principal))
		}
	}

	return schema.NewSet(schema.HashString, vPrincipals)
}
