package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsCloudSearchDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudSearchDomainCreate,
		Read:   resourceAwsCloudSearchDomainRead,
		Update: resourceAwsCloudSearchDomainUpdate,
		Delete: resourceAwsCloudSearchDomainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsCloudSearchDomainImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateDomainName,
			},

			"instance_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "search.small",
				ValidateFunc: validateInstanceType,
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

			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"service_access_policies": {
				Type:             schema.TypeString,
				ValidateFunc:     validateIAMPolicyJson,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"index": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateIndexName,
						},

						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateIndexType,
						},

						"search": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"facet": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"return": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"sort": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"highlight": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

			"wait_for_endpoints": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return true },
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"document_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"search_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

// Terraform CRUD Functions
func resourceAwsCloudSearchDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	input := cloudsearch.CreateDomainInput{
		DomainName: aws.String(d.Get("name").(string)),
	}

	output, err := conn.CreateDomain(&input)
	if err != nil {
		log.Printf("[DEBUG] Creating CloudSearch Domain: %#v", input)
		return fmt.Errorf("%s %q", err, d.Get("name").(string))
	}

	d.SetId(*output.DomainStatus.ARN)
	return resourceAwsCloudSearchDomainUpdate(d, meta)
}

func resourceAwsCloudSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	domainlist := cloudsearch.DescribeDomainsInput{
		DomainNames: []*string{
			aws.String(d.Get("name").(string)),
		},
	}

	resp, err := conn.DescribeDomains(&domainlist)
	if err != nil {
		return err
	}
	domain := resp.DomainStatusList[0]
	d.Set("id", domain.DomainId)

	if domain.DocService.Endpoint != nil {
		d.Set("document_endpoint", domain.DocService.Endpoint)
	}
	if domain.SearchService.Endpoint != nil {
		d.Set("search_endpoint", domain.SearchService.Endpoint)
	}

	// Read index fields.
	indexResults, err := conn.DescribeIndexFields(&cloudsearch.DescribeIndexFieldsInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(indexResults.IndexFields))

	for _, raw := range indexResults.IndexFields {
		// Don't read in any fields that are pending deletion.
		if *raw.Status.PendingDeletion {
			continue
		}

		result = append(result, readIndexField(raw.Options))
	}
	d.Set("index", result)

	// Read service access policies.
	policyResult, err := conn.DescribeServiceAccessPolicies(&cloudsearch.DescribeServiceAccessPoliciesInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}
	d.Set("service_access_policies", policyResult.AccessPolicies.Options)

	// Read availability options (i.e. multi-az).
	availabilityResult, err := conn.DescribeAvailabilityOptions(&cloudsearch.DescribeAvailabilityOptionsInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}
	d.Set("multi_az", availabilityResult.AvailabilityOptions.Options)

	// Read scaling parameters.
	scalingResult, err := conn.DescribeScalingParameters(&cloudsearch.DescribeScalingParametersInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}
	d.Set("instance_type", scalingResult.ScalingParameters.Options.DesiredInstanceType)
	d.Set("partition_count", scalingResult.ScalingParameters.Options.DesiredPartitionCount)
	d.Set("replication_count", scalingResult.ScalingParameters.Options.DesiredReplicationCount)

	return err
}

func resourceAwsCloudSearchDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	_, err := conn.UpdateServiceAccessPolicies(&cloudsearch.UpdateServiceAccessPoliciesInput{
		DomainName:     aws.String(d.Get("name").(string)),
		AccessPolicies: aws.String(d.Get("service_access_policies").(string)),
	})

	_, err = conn.UpdateScalingParameters(&cloudsearch.UpdateScalingParametersInput{
		DomainName: aws.String(d.Get("name").(string)),
		ScalingParameters: &cloudsearch.ScalingParameters{
			DesiredInstanceType:     aws.String(d.Get("instance_type").(string)),
			DesiredReplicationCount: aws.Int64(int64(d.Get("replication_count").(int))),
			DesiredPartitionCount:   aws.Int64(int64(d.Get("partition_count").(int))),
		},
	})

	_, err = conn.UpdateAvailabilityOptions(&cloudsearch.UpdateAvailabilityOptionsInput{
		DomainName: aws.String(d.Get("name").(string)),
		MultiAZ:    aws.Bool(d.Get("multi_az").(bool)),
	})

	updated, err := defineIndexFields(d, meta, conn)
	if err != nil {
		return err
	}

	// When you add fields or modify existing fields, you must explicitly issue a request to re-index your data
	// when you are done making configuration changes.
	// https://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-index-fields.html
	if updated {
		_, err := conn.IndexDocuments(&cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
	}

	if d.Get("wait_for_endpoints").(bool) {
		domainsList := cloudsearch.DescribeDomainsInput{
			DomainNames: []*string{
				aws.String(d.Get("name").(string)),
			},
		}

		err = waitForSearchDomainToBeAvailable(d, conn, domainsList)
		if err != nil {
			return fmt.Errorf("%s %q", err, d.Get("name").(string))
		}
	}
	return nil
}

func resourceAwsCloudSearchDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	dm := d.Get("name").(string)
	input := cloudsearch.DeleteDomainInput{
		DomainName: aws.String(dm),
	}

	_, err := conn.DeleteDomain(&input)

	return err
}

// Import Function
func resourceAwsCloudSearchDomainImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	arn, err := arn.Parse(d.Id())
	if err != nil {
		return nil, err
	}

	parts := strings.Split(arn.Resource, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("resource component of ARN is not properly formatted")
	}

	d.Set("name", parts[1])

	return []*schema.ResourceData{d}, nil
}

// Validation Functions
func validateDomainName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-z]([a-z0-9-]){2,27}$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with a lower-case letter, contain only [a-z0-9-] and be at least 3 and at most 28 characters", k))
	}
	return
}

func validateInstanceType(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	found := false
	for _, t := range cloudsearch.PartitionInstanceType_Values() {
		if t == value {
			found = true
			continue
		}
	}

	if !found {
		es = append(es, fmt.Errorf("%q is not a valid instance type", v))
	}

	return
}

func validateIndexName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if !regexp.MustCompile(`^(\*?[a-z][a-z0-9_]{2,63}|[a-z][a-z0-9_]{2,63}\*?)$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with a letter and be at least 3 and no more than 64 characters long", k))
	}

	if value == "score" {
		es = append(es, fmt.Errorf("'score' is a reserved field name and cannot be used"))
	}

	return
}

func validateIndexType(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	found := false
	for _, t := range cloudsearch.IndexFieldType_Values() {
		if t == value {
			found = true
			continue
		}
	}

	if !found {
		es = append(es, fmt.Errorf("%q is not a valid index type", v))
	}

	return
}

// Waiters
func waitForSearchDomainToBeAvailable(d *schema.ResourceData, conn *cloudsearch.CloudSearch, domainlist cloudsearch.DescribeDomainsInput) error {
	log.Printf("[INFO] cloudsearch (%#v) waiting for domain endpoint. This usually takes 10 minutes.", domainlist.DomainNames[0])
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
			// Not good enough to wait for processing, have to check for search endpoint.
			domain := resp.DomainStatusList[0]
			log.Printf("[DEBUG] GLEN: Domain = %s", domain)
			processing := strconv.FormatBool(*domain.Processing)
			log.Printf("[DEBUG] GLEN: Processing = %s", processing)
			if domain.SearchService.Endpoint != nil {
				log.Printf("[DEBUG] GLEN: type: %T", domain.SearchService.Endpoint)
				log.Printf("[DEBUG] GLEN: SearchServiceEndpoint = %s", *domain.SearchService.Endpoint)
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
	d.Set("id", domain.DomainId)
	d.Set("document_endpoint", domain.DocService.Endpoint)
	d.Set("search_endpoint", domain.SearchService.Endpoint)

	if err != nil {
		return fmt.Errorf("Error waiting for CloudSearch domain (%#v) to finish processing: %s", domainlist.DomainNames[0], err)
	}
	return err
}

// Miscellaneous helper functions
func defineIndexFields(d *schema.ResourceData, meta interface{}, conn *cloudsearch.CloudSearch) (bool, error) {
	// Early return if we don't have a change.
	if !d.HasChange("index") {
		return false, nil
	}

	o, n := d.GetChange("index")

	old := o.(*schema.Set)
	new := n.(*schema.Set)

	// Returns a set of only old fields, to be deleted.
	toDelete := old.Difference(new)
	for _, index := range toDelete.List() {
		v, _ := index.(map[string]interface{})

		_, err := conn.DeleteIndexField(&cloudsearch.DeleteIndexFieldInput{
			DomainName:     aws.String(d.Get("name").(string)),
			IndexFieldName: aws.String(v["name"].(string)),
		})
		if err != nil {
			return true, err
		}
	}

	// Returns a set of only fields that needs to be added or updated (upserted).
	toUpsert := new.Difference(old)
	for _, index := range toUpsert.List() {
		v, _ := index.(map[string]interface{})

		field, err := generateIndexFieldInput(v)
		if err != nil {
			return true, err
		}

		_, err = conn.DefineIndexField(&cloudsearch.DefineIndexFieldInput{
			DomainName: aws.String(d.Get("name").(string)),
			IndexField: field,
		})
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

/*
extractFromMapToType extracts a specific value from map[string]interface{} into an interface of type
expects: map[string]interface{}, string, interface{}
returns: error
*/
func extractFromMapToType(index map[string]interface{}, property string, t interface{}) error {
	v, ok := index[property]
	if !ok {
		return fmt.Errorf("%s is not a valid property of an index", property)
	}

	if "default_value" == property {
		switch t.(type) {
		case *int:
			d, err := strconv.Atoi(v.(string))
			if err != nil {
				return parseError(v.(string), "int")
			}

			reflect.ValueOf(t).Elem().Set(reflect.ValueOf(d))
		case *float64:
			f, err := strconv.ParseFloat(v.(string), 64)
			if err != nil {
				return parseError(v.(string), "double")
			}

			reflect.ValueOf(t).Elem().Set(reflect.ValueOf(f))
		default:
			if v.(string) != "" {
				reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
			}
		}
		return nil
	}

	reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
	return nil
}

var parseError = func(d string, t string) error {
	return fmt.Errorf("can't convert default_value '%s' of type '%s' to int", d, t)
}

func generateIndexFieldInput(index map[string]interface{}) (*cloudsearch.IndexField, error) {
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

	// NOTE: only way I know of to set a default for this field since not all index fields can use it.
	if analysisScheme == "" {
		analysisScheme = "_en_default_"
	}

	switch index["type"] {
	case "date":
		input.DateOptions = &cloudsearch.DateOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			input.DateOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "date-array":
		input.DateArrayOptions = &cloudsearch.DateArrayOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
		}

		if index["default_value"].(string) != "" {
			input.DateArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "double":
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
	case "double-array":
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
	case "int":
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
	case "int-array":
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
	case "latlon":
		input.LatLonOptions = &cloudsearch.LatLonOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			input.LatLonOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "literal":
		input.LiteralOptions = &cloudsearch.LiteralOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			input.LiteralOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "literal-array":
		input.LiteralArrayOptions = &cloudsearch.LiteralArrayOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
		}

		if index["default_value"].(string) != "" {
			input.LiteralArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "text":
		input.TextOptions = &cloudsearch.TextOptions{
			SortEnabled:      aws.Bool(sort),
			ReturnEnabled:    aws.Bool(returnV),
			HighlightEnabled: aws.Bool(highlight),
			AnalysisScheme:   aws.String(analysisScheme),
		}

		if index["default_value"].(string) != "" {
			input.TextOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "text-array":
		input.TextArrayOptions = &cloudsearch.TextArrayOptions{
			ReturnEnabled:    aws.Bool(returnV),
			HighlightEnabled: aws.Bool(highlight),
			AnalysisScheme:   aws.String(analysisScheme),
		}

		if index["default_value"].(string) != "" {
			input.TextArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	default:
		return input, fmt.Errorf("invalid index field type %s", index["type"])
	}

	return input, nil
}

func readIndexField(raw *cloudsearch.IndexField) map[string]interface{} {
	index := map[string]interface{}{
		"name": raw.IndexFieldName,
		"type": raw.IndexFieldType,
	}

	switch *raw.IndexFieldType {
	case "date":
		index["default_value"] = raw.DateOptions.DefaultValue
		index["facet"] = raw.DateOptions.FacetEnabled
		index["return"] = raw.DateOptions.ReturnEnabled
		index["search"] = raw.DateOptions.SearchEnabled
		index["sort"] = raw.DateOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "date-array":
		index["default_value"] = raw.DateArrayOptions.DefaultValue
		index["facet"] = raw.DateArrayOptions.FacetEnabled
		index["return"] = raw.DateArrayOptions.ReturnEnabled
		index["search"] = raw.DateArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "double":
		index["default_value"] = raw.DoubleOptions.DefaultValue
		index["facet"] = raw.DoubleOptions.FacetEnabled
		index["return"] = raw.DoubleOptions.ReturnEnabled
		index["search"] = raw.DoubleOptions.SearchEnabled
		index["sort"] = raw.DoubleOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "double-array":
		index["default_value"] = raw.DoubleArrayOptions.DefaultValue
		index["facet"] = raw.DoubleArrayOptions.FacetEnabled
		index["return"] = raw.DoubleArrayOptions.ReturnEnabled
		index["search"] = raw.DoubleArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "int":
		index["default_value"] = raw.IntOptions.DefaultValue
		index["facet"] = raw.IntOptions.FacetEnabled
		index["return"] = raw.IntOptions.ReturnEnabled
		index["search"] = raw.IntOptions.SearchEnabled
		index["sort"] = raw.IntOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "int-array":
		index["default_value"] = raw.IntArrayOptions.DefaultValue
		index["facet"] = raw.IntArrayOptions.FacetEnabled
		index["return"] = raw.IntArrayOptions.ReturnEnabled
		index["search"] = raw.IntArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "latlon":
		index["default_value"] = raw.LatLonOptions.DefaultValue
		index["facet"] = raw.LatLonOptions.FacetEnabled
		index["return"] = raw.LatLonOptions.ReturnEnabled
		index["search"] = raw.LatLonOptions.SearchEnabled
		index["sort"] = raw.LatLonOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "literal":
		index["default_value"] = raw.LiteralOptions.DefaultValue
		index["facet"] = raw.LiteralOptions.FacetEnabled
		index["return"] = raw.LiteralOptions.ReturnEnabled
		index["search"] = raw.LiteralOptions.SearchEnabled
		index["sort"] = raw.LiteralOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "literal-array":
		index["default_value"] = raw.LiteralArrayOptions.DefaultValue
		index["facet"] = raw.LiteralArrayOptions.FacetEnabled
		index["return"] = raw.LiteralArrayOptions.ReturnEnabled
		index["search"] = raw.LiteralArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "text":
		index["default_value"] = raw.TextOptions.DefaultValue
		index["analysis_scheme"] = raw.TextOptions.AnalysisScheme
		index["highlight"] = raw.TextOptions.HighlightEnabled
		index["return"] = raw.TextOptions.ReturnEnabled
		index["sort"] = raw.TextOptions.SortEnabled

		// Options that aren't valid for this type.
		index["facet"] = false
		index["search"] = false
	case "text-array":
		index["default_value"] = raw.TextArrayOptions.DefaultValue
		index["analysis_scheme"] = raw.TextArrayOptions.AnalysisScheme
		index["highlight"] = raw.TextArrayOptions.HighlightEnabled
		index["return"] = raw.TextArrayOptions.ReturnEnabled

		// Options that aren't valid for this type.
		index["facet"] = false
		index["search"] = false
		index["sort"] = false
	}

	return index
}
