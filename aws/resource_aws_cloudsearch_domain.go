package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

func resourceAwsCloudSearchDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudSearchDomainCreate,
		Read:   resourceAwsCloudSearchDomainRead,
		Update: resourceAwsCloudSearchDomainUpdate,
		Delete: resourceAwsCloudSearchDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDomainNameRegex,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "search.m1.small",
			},
			"replication_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"partition_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"indexes": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"search": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"facet": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"return": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"sort": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"highlight": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"analysis_scheme": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"access_policy": {
				Type:             schema.TypeString,
				ValidateFunc:     validateIAMPolicyJson,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"docservice": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"searchservice": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wait_for_endpoints": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceAwsCloudSearchDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	input := cloudsearch.CreateDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	output, err := conn.CreateDomain(&input)
	if err != nil {
		log.Printf("[DEBUG] Creating CloudWatch Domain: %#v", input)
		return fmt.Errorf("%s %q", err, d.Get("domain_name").(string))
	}

	d.SetId(*output.DomainStatus.ARN)
	return resourceAwsCloudSearchDomainUpdate(d, meta)
}

func resourceAwsCloudSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	domainlist := cloudsearch.DescribeDomainsInput{
		DomainNames: []*string{
			aws.String(d.Get("domain_name").(string)),
		},
	}

	resp, err1 := conn.DescribeDomains(&domainlist)
	if err1 != nil {
		return err1
	}
	domain := resp.DomainStatusList[0]
	d.Set("domain_id", domain.DomainId)

	if domain.DocService.Endpoint != nil {
		d.Set("docservice", domain.DocService.Endpoint)
	}
	if domain.SearchService.Endpoint != nil {
		d.Set("searchservice", domain.SearchService.Endpoint)
	}

	input := cloudsearch.DescribeIndexFieldsInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	_, err := conn.DescribeIndexFields(&input)
	// if err != nil {
	// 	log.Printf("[DEBUG] Reading CloudWatch Index fields: %#v", input)
	// 	return fmt.Errorf("%s %q", err, d.Get("domain_name").(string))
	// }

	return err
}

func waitForSearchDomainToBeAvailable(d *schema.ResourceData, conn *cloudsearch.CloudSearch, domainlist cloudsearch.DescribeDomainsInput) error {
	log.Printf("[INFO] cloudserach (%#v) waiting for SearchDomain endpoint. This usually takes 10 minutes.", domainlist.DomainNames[0])
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Waiting"},
		Target:     []string{"OK"},
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeDomains(&domainlist)
			log.Printf("[DEBUG] Checking %v", domainlist.DomainNames[0])
			if err != nil {
				log.Printf("[ERROR] Could not find domain (%v).  %s", domainlist.DomainNames[0], err)
				return nil, "", err
			}
			// Not good enough to wait for processing, have to check for serach end point.
			domain := resp.DomainStatusList[0]
			log.Printf("[DEBUG] GLEN: Domain = %s", domain)
			processing := strconv.FormatBool(*domain.Processing)
			log.Printf("[DEBUG] GLEN: Processing = %s", processing)
			if domain.SearchService.Endpoint != nil {
				log.Printf("[DEBUG] GLEN: type: %T", domain.SearchService.Endpoint)
				log.Printf("[DEBUG] GLEN: SearchServiceEndPoint = %s", *domain.SearchService.Endpoint)
			}
			if domain.SearchService.Endpoint == nil || *domain.SearchService.Endpoint == "" {
				return resp, "Waiting", nil
			}
			return resp, "OK", nil

		},
	}

	log.Printf("[DEBUG] Waiting for CloudSearch domain to finish processing: %v", domainlist.DomainNames[0])
	_, err := stateConf.WaitForState()

	// Search service was blank.
	resp, err1 := conn.DescribeDomains(&domainlist)
	if err1 != nil {
		return err1
	}
	domain := resp.DomainStatusList[0]
	d.Set("domain_id", domain.DomainId)
	d.Set("docservice", domain.DocService.Endpoint)
	d.Set("searchservice", domain.SearchService.Endpoint)

	if err != nil {
		return fmt.Errorf("Error waiting for CloudSearch domain (%#v) to finish processing: %s", domainlist.DomainNames[0], err)
	}
	return err
}

func resourceAwsCloudSearchDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	err := updateScalingParameters(d, meta, conn)
	if err != nil {
		return err
	}

	updated, err := defineIndexFields(d, meta, conn)
	if err != nil {
		return err
	}

	err = updateAccessPolicy(d, meta, conn)
	if err != nil {
		return err
	}

	// When you add fields or modify existing fields, you must explicitly issue a request to re-index your data
	// when you are done making configuration changes.
	// https://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-index-fields.html
	if updated {
		_, err := conn.IndexDocuments(&cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Get("domain_name").(string)),
		})
		if err != nil {
			return err
		}
	}
	if d.Get("wait_for_endpoints").(bool) {
		domainlist := cloudsearch.DescribeDomainsInput{
			DomainNames: []*string{
				aws.String(d.Get("domain_name").(string)),
			},
		}
		err2 := waitForSearchDomainToBeAvailable(d, conn, domainlist)
		if err2 != nil {
			fmt.Errorf("%s %q", err2, d.Get("domain_name").(string))
		}
	}
	return nil
}

func updateScalingParameters(d *schema.ResourceData, meta interface{}, conn *cloudsearch.CloudSearch) error {
	input := cloudsearch.UpdateScalingParametersInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
		ScalingParameters: &cloudsearch.ScalingParameters{
			DesiredInstanceType:     aws.String(d.Get("instance_type").(string)),
			DesiredReplicationCount: aws.Int64(int64(d.Get("replication_count").(int))),
		},
	}

	if d.Get("instance_type").(string) == "search.m3.2xlarge" {
		input.ScalingParameters.DesiredPartitionCount = aws.Int64(int64(d.Get("partition_count").(int)))
	}

	_, err := conn.UpdateScalingParameters(&input)
	// if err != nil {
	// 	log.Printf("[DEBUG] Updating Scaling Parameters: %#v", input)
	// 	return fmt.Errorf("%s %q", err, d.Get("domain_name").(string))
	// }
	return err
}

func defineIndexFields(d *schema.ResourceData, meta interface{}, conn *cloudsearch.CloudSearch) (bool, error) {
	if d.HasChange("indexes") {
		old := make(map[string]interface{})
		new := make(map[string]interface{})

		o, n := d.GetChange("indexes")

		for _, ot := range o.([]interface{}) {
			os := ot.(map[string]interface{})
			old[os["name"].(string)] = os
		}

		for _, nt := range n.([]interface{}) {
			ns := nt.(map[string]interface{})
			new[ns["name"].(string)] = ns
		}

		// Handle Removal
		for k := range old {
			if _, ok := new[k]; !ok {
				deleteIndexField(d.Get("domain_name").(string), k, conn)
			}
		}

		for _, v := range new {
			// Handle replaces & additions
			err := defineIndexField(d.Get("domain_name").(string), v.(map[string]interface{}), conn)
			if err != nil {
				return true, err
			}
		}
		return true, nil
	}
	return false, nil
}

func defineIndexField(domainName string, index map[string]interface{}, conn *cloudsearch.CloudSearch) error {
	i, err := genIndexFieldInput(index)
	if err != nil {
		return err
	}

	input := cloudsearch.DefineIndexFieldInput{
		DomainName: aws.String(domainName),
		IndexField: i,
	}

	_, err = conn.DefineIndexField(&input)
	return err
}

func deleteIndexField(domainName string, indexName string, conn *cloudsearch.CloudSearch) error {
	input := cloudsearch.DeleteIndexFieldInput{
		DomainName:     aws.String(domainName),
		IndexFieldName: aws.String(indexName),
	}

	_, err := conn.DeleteIndexField(&input)
	return err
}

var parseError = func(d string, t string) error {
	return fmt.Errorf("can't convert default_value '%s' of type '%s' to int", d, t)
}

/*
extractFromMapToType extracts a specific value from map[string]interface{} into an interface of type
expects: map[string]interface{}, string, interface{}
returns: error
*/
func extractFromMapToType(index map[string]interface{}, prop string, t interface{}) error {
	v, ok := index[prop]
	if !ok {
		return fmt.Errorf("%s is not a valid propery of an index", prop)
	}

	if "default_value" == prop {
		switch t.(type) {
		case *int:
			{
				d, err := strconv.Atoi(v.(string))
				if err != nil {
					return parseError(v.(string), "int")
				}

				reflect.ValueOf(t).Elem().Set(reflect.ValueOf(d))
			}
		case *float64:
			{
				f, err := strconv.ParseFloat(v.(string), 64)
				if err != nil {
					return parseError(v.(string), "double")
				}

				reflect.ValueOf(t).Elem().Set(reflect.ValueOf(f))
			}
		default:
			{
				if v.(string) != "" {
					reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
				}
			}
		}
		return nil
	}

	reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
	return nil
}

func genIndexFieldInput(index map[string]interface{}) (*cloudsearch.IndexField, error) {
	input := &cloudsearch.IndexField{
		IndexFieldName: aws.String(index["name"].(string)),
		IndexFieldType: aws.String(index["type"].(string)),
	}

	var facet bool
	var returnV bool
	var search bool
	var sort bool
	var highlight bool
	var analysisScheme string

	extractFromMapToType(index, "facet", &facet)
	extractFromMapToType(index, "return", &returnV)
	extractFromMapToType(index, "search", &search)
	extractFromMapToType(index, "sort", &sort)
	extractFromMapToType(index, "highlight", &highlight)
	extractFromMapToType(index, "analysis_scheme", &analysisScheme)

	switch index["type"] {
	case "int":
		{
			input.IntOptions = &cloudsearch.IntOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				var defaultValue int
				extractFromMapToType(index, "default_value", &defaultValue)
				input.IntOptions.DefaultValue = aws.Int64(int64(defaultValue))
			}
		}
	case "int-array":
		{
			input.IntArrayOptions = &cloudsearch.IntArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				var defaultValue int
				extractFromMapToType(index, "default_value", &defaultValue)
				input.IntArrayOptions.DefaultValue = aws.Int64(int64(defaultValue))
			}
		}
	case "double":
		{
			input.DoubleOptions = &cloudsearch.DoubleOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				var defaultValue float64
				extractFromMapToType(index, "default_value", &defaultValue)
				input.DoubleOptions.DefaultValue = aws.Float64(float64(defaultValue))
			}
		}
	case "double-array":
		{
			input.DoubleArrayOptions = &cloudsearch.DoubleArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				var defaultValue float64
				extractFromMapToType(index, "default_value", &defaultValue)
				input.DoubleArrayOptions.DefaultValue = aws.Float64(float64(defaultValue))
			}
		}
	case "literal":
		{
			input.LiteralOptions = &cloudsearch.LiteralOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				input.LiteralOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "literal-array":
		{
			input.LiteralArrayOptions = &cloudsearch.LiteralArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				input.LiteralArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "text":
		{
			input.TextOptions = &cloudsearch.TextOptions{
				SortEnabled:      aws.Bool(sort),
				ReturnEnabled:    aws.Bool(returnV),
				HighlightEnabled: aws.Bool(highlight),
				AnalysisScheme:   aws.String(analysisScheme),
			}

			if index["default_value"].(string) != "" {
				input.TextOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "text-array":
		{
			input.TextArrayOptions = &cloudsearch.TextArrayOptions{
				ReturnEnabled:    aws.Bool(returnV),
				HighlightEnabled: aws.Bool(highlight),
				AnalysisScheme:   aws.String(analysisScheme),
			}

			if index["default_value"].(string) != "" {
				input.TextArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "date":
		{
			input.DateOptions = &cloudsearch.DateOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				input.DateOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "date-aray":
		{
			input.DateArrayOptions = &cloudsearch.DateArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				input.DateArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "latlon":
		{
			input.LatLonOptions = &cloudsearch.LatLonOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				input.LatLonOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	default:
		return input, fmt.Errorf("invalid index field type %s", index["type"])
	}

	return input, nil
}

func updateAccessPolicy(d *schema.ResourceData, meta interface{}, conn *cloudsearch.CloudSearch) error {
	input := cloudsearch.UpdateServiceAccessPoliciesInput{
		DomainName:     aws.String(d.Get("domain_name").(string)),
		AccessPolicies: aws.String(d.Get("access_policy").(string)),
	}

	_, err := conn.UpdateServiceAccessPolicies(&input)
	return err
}

func resourceAwsCloudSearchDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	dm := d.Get("domain_name").(string)
	input := cloudsearch.DeleteDomainInput{
		DomainName: aws.String(dm),
	}

	_, err := conn.DeleteDomain(&input)

	return err
}

func validateDomainNameRegex(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-z]([a-z0-9-]){2,27}$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with a lower-case letter, contain only [a-z0-9-] and be at least 3 and at most 28 characters", k))
	}
	return
}
