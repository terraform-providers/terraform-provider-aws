package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
	"sort"
	"time"
)

func resourceAwsAppstreamFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppstreamFleetCreate,
		Read:   resourceAwsAppstreamFleetRead,
		Update: resourceAwsAppstreamFleetUpdate,
		Delete: resourceAwsAppstreamFleetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"compute_capacity": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_instances": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"disconnect_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"domain_info": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"organizational_unit_distinguished_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"enable_default_internet_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"fleet_type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"image_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"max_user_duration": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"stack_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"state": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAppstreamFleetCreate(d *schema.ResourceData, meta interface{}) error {

	svc := meta.(*AWSClient).appstreamconn
	CreateFleetInputOpts := &appstream.CreateFleetInput{}

	if v, ok := d.GetOk("name"); ok {
		CreateFleetInputOpts.Name = aws.String(v.(string))
	}

	ComputeConfig := &appstream.ComputeCapacity{}

	if a, ok := d.GetOk("compute_capacity"); ok {
		ComputeAttributes := a.([]interface{})
		attr := ComputeAttributes[0].(map[string]interface{})
		if v, ok := attr["desired_instances"]; ok {
			ComputeConfig.DesiredInstances = aws.Int64(int64(v.(int)))
		}
		CreateFleetInputOpts.ComputeCapacity = ComputeConfig
	}

	if v, ok := d.GetOk("description"); ok {
		CreateFleetInputOpts.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disconnect_timeout"); ok {
		CreateFleetInputOpts.DisconnectTimeoutInSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("display_name"); ok {
		CreateFleetInputOpts.DisplayName = aws.String(v.(string))
	}

	DomainJoinInfoConfig := &appstream.DomainJoinInfo{}

	if dom, ok := d.GetOk("domain_info"); ok {
		DomainAttributes := dom.([]interface{})
		attr := DomainAttributes[0].(map[string]interface{})
		if v, ok := attr["directory_name"]; ok {
			DomainJoinInfoConfig.DirectoryName = aws.String(v.(string))
		}
		if v, ok := attr["organizational_unit_distinguished_name"]; ok {
			DomainJoinInfoConfig.OrganizationalUnitDistinguishedName = aws.String(v.(string))
		}
		CreateFleetInputOpts.DomainJoinInfo = DomainJoinInfoConfig
	}

	if v, ok := d.GetOk("enable_default_internet_access"); ok {
		CreateFleetInputOpts.EnableDefaultInternetAccess = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("fleet_type"); ok {
		CreateFleetInputOpts.FleetType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_name"); ok {
		CreateFleetInputOpts.ImageName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_type"); ok {
		CreateFleetInputOpts.InstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_user_duration"); ok {
		CreateFleetInputOpts.MaxUserDurationInSeconds = aws.Int64(int64(v.(int)))
	}

	VpcConfigConfig := &appstream.VpcConfig{}

	if v, ok := d.GetOk("security_group_ids"); ok {
		convertedSecurityGroups := convertToStringSlice(v)
		VpcConfigConfig.SecurityGroupIds = aws.StringSlice(convertedSecurityGroups)
	}
	if v, ok := d.GetOk("subnet_ids"); ok {
		convertedSubnets := convertToStringSlice(v)
		VpcConfigConfig.SubnetIds = aws.StringSlice(convertedSubnets)
	}
	CreateFleetInputOpts.VpcConfig = VpcConfigConfig

	log.Printf("[DEBUG] Run configuration: %s", CreateFleetInputOpts)
	resp, err := svc.CreateFleet(CreateFleetInputOpts)

	if err != nil {
		log.Printf("[ERROR] Error creating Appstream Fleet: %s", err)
		return err
	}

	log.Printf("[DEBUG] %s", resp)
	time.Sleep(2 * time.Second)
	if v, ok := d.GetOk("tags"); ok {

		data_tags := v.(map[string]interface{})

		attr := make(map[string]string)

		for k, v := range data_tags {
			attr[k] = v.(string)
		}

		tags := aws.StringMap(attr)

		fleet_name := aws.StringValue(CreateFleetInputOpts.Name)
		get, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{
			Names: aws.StringSlice([]string{fleet_name}),
		})
		if err != nil {
			log.Printf("[ERROR] Error describing Appstream Fleet: %s", err)
			return err
		}
		if get.Fleets == nil {
			log.Printf("[DEBUG] Apsstream Fleet (%s) not found", d.Id())
		}

		fleetArn := get.Fleets[0].Arn

		tag, err := svc.TagResource(&appstream.TagResourceInput{
			ResourceArn: fleetArn,
			Tags:        tags,
		})
		if err != nil {
			log.Printf("[ERROR] Error tagging Appstream Stack: %s", err)
			return err
		}
		log.Printf("[DEBUG] %s", tag)
	}

	if v, ok := d.GetOk("stack_name"); ok {
		AssociateFleetInputOpts := &appstream.AssociateFleetInput{}
		AssociateFleetInputOpts.FleetName = CreateFleetInputOpts.Name
		AssociateFleetInputOpts.StackName = aws.String(v.(string))
		resp, err := svc.AssociateFleet(AssociateFleetInputOpts)
		if err != nil {
			log.Printf("[ERROR] Error associating Appstream Fleet: %s", err)
			return err
		}

		log.Printf("[DEBUG] %s", resp)
	}

	if v, ok := d.GetOk("state"); ok {
		if v == "RUNNING" {
			desired_state := v
			resp, err := svc.StartFleet(&appstream.StartFleetInput{
				Name: CreateFleetInputOpts.Name,
			})

			if err != nil {
				log.Printf("[ERROR] Error starting Appstream Fleet: %s", err)
				return err
			}
			log.Printf("[DEBUG] %s", resp)

			for {

				resp, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{
					Names: aws.StringSlice([]string{*CreateFleetInputOpts.Name}),
				})

				if err != nil {
					log.Printf("[ERROR] Error describing Appstream Fleet: %s", err)
					return err
				}

				curr_state := resp.Fleets[0].State
				if aws.StringValue(curr_state) == desired_state {
					break
				}
				if aws.StringValue(curr_state) != desired_state {
					time.Sleep(20 * time.Second)
					continue
				}

			}
		}
	}

	d.SetId(*CreateFleetInputOpts.Name)

	return resourceAwsAppstreamFleetRead(d, meta)
}

func resourceAwsAppstreamFleetRead(d *schema.ResourceData, meta interface{}) error {
	svc := meta.(*AWSClient).appstreamconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{})

	if err != nil {
		log.Printf("[ERROR] Error reading Appstream Fleet: %s", err)
		return err
	}
	for _, v := range resp.Fleets {
		if aws.StringValue(v.Name) == d.Get("name") {

			d.Set("name", v.Name)

			if v.ComputeCapacityStatus != nil {
				comp_attr := map[string]interface{}{}
				comp_attr["desired_instances"] = aws.Int64Value(v.ComputeCapacityStatus.Desired)
				d.Set("compute_capacity", comp_attr)
			}

			d.Set("description", v.Description)
			d.Set("display_name", v.DisplayName)
			d.Set("disconnect_timeout", v.DisconnectTimeoutInSeconds)
			d.Set("enable_default_internet_access", v.EnableDefaultInternetAccess)
			d.Set("fleet_type", v.FleetType)
			d.Set("image_name", v.ImageName)
			d.Set("instance_type", v.InstanceType)
			d.Set("max_user_duration", v.MaxUserDurationInSeconds)

			if v.VpcConfig != nil {
				d.Set("security_group_ids", aws.StringValueSlice(v.VpcConfig.SecurityGroupIds))
				d.Set("subnet_ids", aws.StringValueSlice(v.VpcConfig.SubnetIds))
			}
			tg, err := svc.ListTagsForResource(&appstream.ListTagsForResourceInput{
				ResourceArn: v.Arn,
			})

			if err != nil {
				log.Printf("[ERROR] Error listing stack tags: %s", err)
				return err
			}
			if tg.Tags == nil {
				log.Printf("[DEBUG] Apsstream Stack tags (%s) not found", d.Id())
				return nil
			}
			d.Set("tags", keyvaluetags.AppstreamKeyValueTags(tg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())

			d.Set("state", v.State)

			return nil

		}
	}
	d.SetId("")
	return nil
}

func resourceAwsAppstreamFleetUpdate(d *schema.ResourceData, meta interface{}) error {

	svc := meta.(*AWSClient).appstreamconn
	UpdateFleetInputOpts := &appstream.UpdateFleetInput{}

	if v, ok := d.GetOk("name"); ok {
		UpdateFleetInputOpts.Name = aws.String(v.(string))
	}

	if d.HasChange("description") {
		log.Printf("[DEBUG] Modify Fleet")
		description := d.Get("description").(string)
		UpdateFleetInputOpts.Description = aws.String(description)
	}

	if d.HasChange("disconnect_timeout") {
		log.Printf("[DEBUG] Modify Fleet")
		disconnect_timeout := d.Get("disconnect_timeout").(int)
		UpdateFleetInputOpts.DisconnectTimeoutInSeconds = aws.Int64(int64(disconnect_timeout))
	}

	if d.HasChange("display_name") {
		log.Printf("[DEBUG] Modify Fleet")
		display_name := d.Get("display_name").(string)
		UpdateFleetInputOpts.DisplayName = aws.String(display_name)
	}

	if d.HasChange("image_name") {
		log.Printf("[DEBUG] Modify Fleet")
		image_name := d.Get("image_name").(string)
		UpdateFleetInputOpts.ImageName = aws.String(image_name)
	}

	if d.HasChange("instance_type") {
		log.Printf("[DEBUG] Modify Fleet")
		instance_type := d.Get("instance_type").(string)
		UpdateFleetInputOpts.InstanceType = aws.String(instance_type)
	}

	if d.HasChange("max_user_duration") {
		log.Printf("[DEBUG] Modify Fleet")
		max_user_duration := d.Get("max_user_duration").(int)
		UpdateFleetInputOpts.MaxUserDurationInSeconds = aws.Int64(int64(max_user_duration))
	}

	if d.HasChanges("security_group_ids", "subnet_ids") {
		log.Printf("[DEBUG] Modify Fleet")
		vpcConfig := &appstream.VpcConfig{}
		convertedSubnets := convertToStringSlice(d.Get("subnet_ids"))
		vpcConfig.SubnetIds = aws.StringSlice(convertedSubnets)
		convertedSecurityGroups := convertToStringSlice(d.Get("security_group_ids"))
		vpcConfig.SecurityGroupIds = aws.StringSlice(convertedSecurityGroups)
		UpdateFleetInputOpts.VpcConfig = vpcConfig
	}

	resp, err := svc.UpdateFleet(UpdateFleetInputOpts)
	if err != nil {
		log.Printf("[ERROR] Error updating Appstream Fleet: %s", err)
		return err
	}

	if d.HasChange("tags") {
		arn := aws.StringValue(resp.Fleet.Arn)

		o, n := d.GetChange("tags")
		if err := keyvaluetags.AppstreamUpdateTags(svc, arn, o, n); err != nil {
			log.Printf("error updating AppStream fleet (%s) tags: %s", d.Id(), err)
			return err
		}
	}

	log.Printf("[DEBUG] %s", resp)
	desired_state := d.Get("state")
	if d.HasChange("state") {
		if desired_state == "STOPPED" {
			_, err := svc.StopFleet(&appstream.StopFleetInput{
				Name: aws.String(d.Id()),
			})
			if err != nil {
				return err
			}
			for {

				resp, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{
					Names: aws.StringSlice([]string{*UpdateFleetInputOpts.Name}),
				})
				if err != nil {
					log.Printf("[ERROR] Error describing Appstream Fleet: %s", err)
					return err
				}

				curr_state := resp.Fleets[0].State
				if aws.StringValue(curr_state) == desired_state {
					break
				}
				if aws.StringValue(curr_state) != desired_state {
					time.Sleep(20 * time.Second)
					continue
				}
			}
		} else if desired_state == "RUNNING" {
			_, err := svc.StartFleet(&appstream.StartFleetInput{
				Name: aws.String(d.Id()),
			})
			if err != nil {
				return err
			}
			for {

				resp, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{
					Names: aws.StringSlice([]string{*UpdateFleetInputOpts.Name}),
				})
				if err != nil {
					log.Printf("[ERROR] Error describing Appstream Fleet: %s", err)
					return err
				}

				curr_state := resp.Fleets[0].State
				if aws.StringValue(curr_state) == desired_state {
					break
				}
				if aws.StringValue(curr_state) != desired_state {
					time.Sleep(20 * time.Second)
					continue
				}

			}
		}
	}
	return resourceAwsAppstreamFleetRead(d, meta)

}

func resourceAwsAppstreamFleetDelete(d *schema.ResourceData, meta interface{}) error {

	svc := meta.(*AWSClient).appstreamconn

	resp, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{
		Names: aws.StringSlice([]string{*aws.String(d.Id())}),
	})

	if err != nil {
		log.Printf("[ERROR] Error reading Appstream Fleet: %s", err)
		return err
	}

	curr_state := aws.StringValue(resp.Fleets[0].State)

	if curr_state == "RUNNING" {
		desired_state := "STOPPED"
		_, err := svc.StopFleet(&appstream.StopFleetInput{
			Name: aws.String(d.Id()),
		})
		if err != nil {
			return err
		}
		for {

			resp, err := svc.DescribeFleets(&appstream.DescribeFleetsInput{
				Names: aws.StringSlice([]string{*aws.String(d.Id())}),
			})
			if err != nil {
				log.Printf("[ERROR] Error describing Appstream Fleet: %s", err)
				return err
			}

			curr_state := resp.Fleets[0].State
			if aws.StringValue(curr_state) == desired_state {
				break
			}
			if aws.StringValue(curr_state) != desired_state {
				time.Sleep(20 * time.Second)
				continue
			}

		}

	}

	dis, err := svc.DisassociateFleet(&appstream.DisassociateFleetInput{
		FleetName: aws.String(d.Id()),
		StackName: aws.String(d.Get("stack_name").(string)),
	})
	if err != nil {
		log.Printf("[ERROR] Error deleting Appstream Fleet: %s", err)
		return err
	}
	log.Printf("[DEBUG] %s", dis)

	del, err := svc.DeleteFleet(&appstream.DeleteFleetInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		log.Printf("[ERROR] Error deleting Appstream Fleet: %s", err)
		return err
	}
	log.Printf("[DEBUG] %s", del)
	return nil

}

func convertToStringSlice(v interface{}) []string {
	// schema object for TypeList is a list of interfaces, each element
	// needs to be converted to string
	// then sorted to keep hash code consistent.
	//
	// code borrowed from resource_aws_security_group.go
	vs := v.([]interface{})
	s := make([]string, len(vs))
	for i, raw := range vs {
		s[i] = raw.(string)
	}
	sort.Strings(s)

	return s
}
