package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func resourceAwsRoute53TrafficPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53TrafficPolicyCreate,
		Read:   resourceAwsRoute53TrafficPolicyRead,
		Update: resourceAwsRoute53TrafficPolicyUpdate,
		Delete: resourceAwsRoute53TrafficPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected traffic-policy-id/traffic-policy-version", d.Id())
				}
				version, err := strconv.Atoi(idParts[1])
				if err != nil {
					return nil, fmt.Errorf("Cannot convert to int: %s", idParts[1])
				}
				d.Set("latest_version", version)
				d.SetId(idParts[0])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"comment": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"document": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 102400),
			},
			"latest_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53TrafficPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	request := &route53.CreateTrafficPolicyInput{
		Name:     aws.String(d.Get("name").(string)),
		Comment:  aws.String(d.Get("comment").(string)),
		Document: aws.String(d.Get("document").(string)),
	}

	response, err := conn.CreateTrafficPolicy(request)
	if err != nil {
		return fmt.Errorf("Error creating Route53 Traffic Policy %s: %s", d.Get("name").(string), err)
	}

	d.SetId(*response.TrafficPolicy.Id)

	err = d.Set("latest_version", response.TrafficPolicy.Version)
	if err != nil {
		return fmt.Errorf("Error assigning Id for Route53 Traffic Policy %s: %s", d.Get("name").(string), err)
	}

	return resourceAwsRoute53TrafficPolicyRead(d, meta)
}

func resourceAwsRoute53TrafficPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	tp, err := getTrafficPolicyById(d.Id(), conn)
	if err != nil {
		return err
	}

	if tp == nil {
		log.Printf("[WARN] Route53 Traffic Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	request := &route53.GetTrafficPolicyInput{
		Id:      tp.Id,
		Version: tp.LatestVersion,
	}

	response, err := conn.GetTrafficPolicy(request)
	if err != nil {
		return fmt.Errorf("Error getting Route53 Traffic Policy %s, version %d: %s", d.Get("name").(string), d.Get("latest_version").(int), err)
	}

	err = d.Set("document", response.TrafficPolicy.Document)
	if err != nil {
		return fmt.Errorf("Error setting document for: %s, error: %#v", d.Id(), err)
	}

	err = d.Set("name", response.TrafficPolicy.Name)
	if err != nil {
		return fmt.Errorf("Error setting name for: %s, error: %#v", d.Id(), err)
	}

	err = d.Set("comment", response.TrafficPolicy.Comment)
	if err != nil {
		return fmt.Errorf("Error setting comment for: %s, error: %#v", d.Id(), err)
	}

	return nil
}

func getTrafficPolicyById(trafficPolicyId string, conn *route53.Route53) (*route53.TrafficPolicySummary, error) {
	var idMarker *string

	for allPoliciesListed := false; !allPoliciesListed; {
		listRequest := &route53.ListTrafficPoliciesInput{}

		if idMarker != nil {
			listRequest.TrafficPolicyIdMarker = idMarker
		}

		listResponse, err := conn.ListTrafficPolicies(listRequest)
		if err != nil {
			return nil, fmt.Errorf("Error listing Route 53 Traffic Policies: %v", err)
		}

		for _, tp := range listResponse.TrafficPolicySummaries {
			if *tp.Id == trafficPolicyId {
				return tp, nil
			}
		}

		if *listResponse.IsTruncated {
			idMarker = listResponse.TrafficPolicyIdMarker
		} else {
			allPoliciesListed = true
		}
	}
	return nil, nil
}

func resourceAwsRoute53TrafficPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	request := &route53.CreateTrafficPolicyVersionInput{
		Id:       aws.String(d.Id()),
		Comment:  aws.String(d.Get("comment").(string)),
		Document: aws.String(d.Get("document").(string)),
	}

	response, err := conn.CreateTrafficPolicyVersion(request)
	if err != nil {
		return fmt.Errorf("Error updating Route53 Traffic Policy: %s. %#v", d.Get("name").(string), err)
	}

	err = d.Set("latest_version", response.TrafficPolicy.Version)
	if err != nil {
		return fmt.Errorf("Error updating Route53 Traffic Policy %s and setting new policy version. %#v", d.Get("name").(string), err)
	}

	return resourceAwsRoute53TrafficPolicyRead(d, meta)
}

func resourceAwsRoute53TrafficPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	var versionMarker *string

	var trafficPolicies []*route53.TrafficPolicy

	for allPoliciesListed := false; !allPoliciesListed; {
		listRequest := &route53.ListTrafficPolicyVersionsInput{
			Id: aws.String(d.Id()),
		}
		if versionMarker != nil {
			listRequest.TrafficPolicyVersionMarker = versionMarker
		}

		listResponse, err := conn.ListTrafficPolicyVersions(listRequest)
		if err != nil {
			return fmt.Errorf("Error listing Route 53 Traffic Policy versions: %v", err)
		}

		trafficPolicies = append(trafficPolicies, listResponse.TrafficPolicies...)

		if *listResponse.IsTruncated {
			versionMarker = listResponse.TrafficPolicyVersionMarker
		} else {
			allPoliciesListed = true
		}
	}

	for _, trafficPolicy := range trafficPolicies {
		deleteRequest := &route53.DeleteTrafficPolicyInput{
			Id:      trafficPolicy.Id,
			Version: trafficPolicy.Version,
		}

		_, err := conn.DeleteTrafficPolicy(deleteRequest)
		if err != nil {
			return fmt.Errorf("Error deleting Route53 Traffic Policy %s, version %d: %s", *trafficPolicy.Id, *trafficPolicy.Version, err)
		}
	}

	return nil
}
