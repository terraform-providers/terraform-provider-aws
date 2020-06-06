package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsLbbListenerRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbListenerRuleCreate,
		Read:   resourceAwsLbListenerRuleRead,
		Update: resourceAwsLbListenerRuleUpdate,
		Delete: resourceAwsLbListenerRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsLbListenerRulePriority,
			},
			"action": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								elbv2.ActionTypeEnumAuthenticateCognito,
								elbv2.ActionTypeEnumAuthenticateOidc,
								elbv2.ActionTypeEnumFixedResponse,
								elbv2.ActionTypeEnumForward,
								elbv2.ActionTypeEnumRedirect,
							}, true),
						},
						"order": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 50000),
						},

						"target_group_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(elbv2.ActionTypeEnumForward),
						},

						"forward": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(elbv2.ActionTypeEnumForward),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_group": {
										Type:     schema.TypeSet,
										MinItems: 2,
										MaxItems: 5,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"arn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"weight": {
													Type:         schema.TypeInt,
													ValidateFunc: validation.IntBetween(0, 999),
													Default:      1,
													Optional:     true,
												},
											},
										},
									},
									"stickiness": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												"duration": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 604800),
												},
											},
										},
									},
								},
							},
						},

						"redirect": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(elbv2.ActionTypeEnumRedirect),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{host}",
									},

									"path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/#{path}",
									},

									"port": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{port}",
									},

									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{protocol}",
										ValidateFunc: validation.StringInSlice([]string{
											"#{protocol}",
											"HTTP",
											"HTTPS",
										}, false),
									},

									"query": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "#{query}",
									},

									"status_code": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"HTTP_301",
											"HTTP_302",
										}, false),
									},
								},
							},
						},

						"fixed_response": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(elbv2.ActionTypeEnumFixedResponse),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"text/plain",
											"text/css",
											"text/html",
											"application/javascript",
											"application/json",
										}, false),
									},

									"message_body": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"status_code": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[245]\d\d$`), ""),
									},
								},
							},
						},

						"authenticate_cognito": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(elbv2.ActionTypeEnumAuthenticateCognito),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice([]string{
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumDeny,
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumAllow,
											elbv2.AuthenticateCognitoActionConditionalBehaviorEnumAuthenticate,
										}, true),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"user_pool_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool_client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_pool_domain": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},

						"authenticate_oidc": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: suppressIfActionTypeNot(elbv2.ActionTypeEnumAuthenticateOidc),
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_request_extra_params": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"authorization_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"client_secret": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Required: true,
									},
									"on_unauthenticated_request": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.StringInSlice([]string{
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumDeny,
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumAllow,
											elbv2.AuthenticateOidcActionConditionalBehaviorEnumAuthenticate,
										}, true),
									},
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_cookie_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"session_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"token_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_info_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"condition": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      lbListenerRuleConditionSetHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"host-header",
								"path-pattern",
							}, true),
							Deprecated: "use 'host_header' or 'path_pattern' attribute instead",
						},
						"host_header": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true, // Deprecated: remove Computed
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type: schema.TypeSet,
										// Deprecated: Change Optional & Computed to Required in next major version of the provider
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Set: schema.HashString,
									},
								},
							},
						},
						"http_header": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_header_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile("^[A-Za-z0-9!#$%&'*+-.^_`|~]{1,40}$"), ""),
									},
									"values": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Required: true,
										Set:      schema.HashString,
									},
								},
							},
						},
						"http_request_method": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z-_]{1,40}$`), ""),
										},
										Required: true,
										Set:      schema.HashString,
									},
								},
							},
						},
						"path_pattern": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true, // Deprecated: remove Computed
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type: schema.TypeSet,
										// Deprecated: Change Optional & Computed to Required in next major version of the provider
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
										Set: schema.HashString,
									},
								},
							},
						},
						"query_string": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							Set: lbListenerRuleConditionQueryStringHash,
						},
						"source_ip": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"values": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateCIDRNetworkAddress,
										},
										Required: true,
										Set:      schema.HashString,
									},
								},
							},
						},
						"values": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 128),
							},
							Optional:   true,
							Computed:   true,
							Deprecated: "use 'host_header' or 'path_pattern' attribute instead",
						},
					},
				},
			},
		},
	}
}

/* DEPRECATED Backwards compatibility: This primarily exists to set a hash that handles the values to host_header or path_pattern migration.
Can probably be removed on the next major version of the provider.
*/
func lbListenerRuleConditionSetHash(v interface{}) int {
	if v == nil {
		return 0
	}

	var field string
	var buf strings.Builder

	m := v.(map[string]interface{})

	if hostHeader, ok := m["host_header"].([]interface{}); ok && len(hostHeader) > 0 {
		if hostHeader[0] != nil {
			field = "host-header"
			values := hostHeader[0].(map[string]interface{})["values"].(*schema.Set)
			for _, l := range values.List() {
				fmt.Fprint(&buf, l, "-")
			}
		}
	} else if m["field"].(string) == "host-header" {
		// Backwards compatibility
		field = "host-header"
		for _, l := range m["values"].([]interface{}) {
			fmt.Fprint(&buf, l, "-")
		}
	}

	if httpHeader, ok := m["http_header"].([]interface{}); ok && len(httpHeader) > 0 && httpHeader[0] != nil {
		field = "http-header"
		httpHeaderMap := httpHeader[0].(map[string]interface{})
		fmt.Fprint(&buf, httpHeaderMap["http_header_name"].(string), ":")
		httpHeaderValues := httpHeaderMap["values"].(*schema.Set)
		for _, l := range httpHeaderValues.List() {
			fmt.Fprint(&buf, l, "-")
		}
	}

	if httpRequestMethod, ok := m["http_request_method"].([]interface{}); ok && len(httpRequestMethod) > 0 && httpRequestMethod[0] != nil {
		field = "http-request-method"
		values := httpRequestMethod[0].(map[string]interface{})["values"].(*schema.Set)
		for _, l := range values.List() {
			fmt.Fprint(&buf, l, "-")
		}
	}

	if pathPattern, ok := m["path_pattern"].([]interface{}); ok && len(pathPattern) > 0 {
		if pathPattern[0] != nil {
			field = "path-pattern"
			values := pathPattern[0].(map[string]interface{})["values"].(*schema.Set)
			for _, l := range values.List() {
				fmt.Fprint(&buf, l, "-")
			}
		}
	} else if m["field"].(string) == "path-pattern" {
		// Backwards compatibility
		field = "path-pattern"
		for _, l := range m["values"].([]interface{}) {
			fmt.Fprint(&buf, l, "-")
		}
	}

	if queryString, ok := m["query_string"].(*schema.Set); ok && queryString.Len() > 0 {
		field = "query-string"
		for _, l := range queryString.List() {
			fmt.Fprint(&buf, lbListenerRuleConditionQueryStringHash(l), "-")
		}
	}

	if sourceIp, ok := m["source_ip"].([]interface{}); ok && len(sourceIp) > 0 && sourceIp[0] != nil {
		field = "source-ip"
		values := sourceIp[0].(map[string]interface{})["values"].(*schema.Set)
		for _, l := range values.List() {
			fmt.Fprint(&buf, l, "-")
		}
	}

	return hashcode.String(fmt.Sprintf("%s-%s", field, buf.String()))
}

func lbListenerRuleConditionQueryStringHash(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf("%s-%s", m["key"], m["value"]))
}

func suppressIfActionTypeNot(t string) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		take := 2
		i := strings.IndexFunc(k, func(r rune) bool {
			if r == '.' {
				take -= 1
				return take == 0
			}
			return false
		})
		at := k[:i+1] + "type"
		return d.Get(at).(string) != t
	}
}

func resourceAwsLbListenerRuleCreate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn
	listenerArn := d.Get("listener_arn").(string)

	params := &elbv2.CreateRuleInput{
		ListenerArn: aws.String(listenerArn),
	}

	actions := d.Get("action").([]interface{})
	params.Actions = make([]*elbv2.Action, len(actions))
	for i, action := range actions {
		actionMap := action.(map[string]interface{})

		action := &elbv2.Action{
			Order: aws.Int64(int64(i + 1)),
			Type:  aws.String(actionMap["type"].(string)),
		}

		if order, ok := actionMap["order"]; ok && order.(int) != 0 {
			action.Order = aws.Int64(int64(order.(int)))
		}

		switch actionMap["type"].(string) {
		case elbv2.ActionTypeEnumForward:
			if err := lbListenerRuleActionForward(actionMap, action); err != nil {
				return err
			}

		case elbv2.ActionTypeEnumRedirect:
			redirectList := actionMap["redirect"].([]interface{})

			if len(redirectList) == 1 {
				redirectMap := redirectList[0].(map[string]interface{})

				action.RedirectConfig = &elbv2.RedirectActionConfig{
					Host:       aws.String(redirectMap["host"].(string)),
					Path:       aws.String(redirectMap["path"].(string)),
					Port:       aws.String(redirectMap["port"].(string)),
					Protocol:   aws.String(redirectMap["protocol"].(string)),
					Query:      aws.String(redirectMap["query"].(string)),
					StatusCode: aws.String(redirectMap["status_code"].(string)),
				}
			} else {
				return errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
			}

		case elbv2.ActionTypeEnumFixedResponse:
			fixedResponseList := actionMap["fixed_response"].([]interface{})

			if len(fixedResponseList) == 1 {
				fixedResponseMap := fixedResponseList[0].(map[string]interface{})

				action.FixedResponseConfig = &elbv2.FixedResponseActionConfig{
					ContentType: aws.String(fixedResponseMap["content_type"].(string)),
					MessageBody: aws.String(fixedResponseMap["message_body"].(string)),
					StatusCode:  aws.String(fixedResponseMap["status_code"].(string)),
				}
			} else {
				return errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
			}

		case elbv2.ActionTypeEnumAuthenticateCognito:
			authenticateCognitoList := actionMap["authenticate_cognito"].([]interface{})

			if len(authenticateCognitoList) == 1 {
				authenticateCognitoMap := authenticateCognitoList[0].(map[string]interface{})

				authenticationRequestExtraParams := make(map[string]*string)
				for key, value := range authenticateCognitoMap["authentication_request_extra_params"].(map[string]interface{}) {
					authenticationRequestExtraParams[key] = aws.String(value.(string))
				}

				action.AuthenticateCognitoConfig = &elbv2.AuthenticateCognitoActionConfig{
					AuthenticationRequestExtraParams: authenticationRequestExtraParams,
					UserPoolArn:                      aws.String(authenticateCognitoMap["user_pool_arn"].(string)),
					UserPoolClientId:                 aws.String(authenticateCognitoMap["user_pool_client_id"].(string)),
					UserPoolDomain:                   aws.String(authenticateCognitoMap["user_pool_domain"].(string)),
				}

				if onUnauthenticatedRequest, ok := authenticateCognitoMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
					action.AuthenticateCognitoConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
				}
				if scope, ok := authenticateCognitoMap["scope"]; ok && scope != "" {
					action.AuthenticateCognitoConfig.Scope = aws.String(scope.(string))
				}
				if sessionCookieName, ok := authenticateCognitoMap["session_cookie_name"]; ok && sessionCookieName != "" {
					action.AuthenticateCognitoConfig.SessionCookieName = aws.String(sessionCookieName.(string))
				}
				if sessionTimeout, ok := authenticateCognitoMap["session_timeout"]; ok && sessionTimeout != 0 {
					action.AuthenticateCognitoConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
				}
			} else {
				return errors.New("for actions of type 'authenticate-cognito', you must specify a 'authenticate_cognito' block")
			}

		case elbv2.ActionTypeEnumAuthenticateOidc:
			authenticateOidcList := actionMap["authenticate_oidc"].([]interface{})

			if len(authenticateOidcList) == 1 {
				authenticateOidcMap := authenticateOidcList[0].(map[string]interface{})

				authenticationRequestExtraParams := make(map[string]*string)
				for key, value := range authenticateOidcMap["authentication_request_extra_params"].(map[string]interface{}) {
					authenticationRequestExtraParams[key] = aws.String(value.(string))
				}

				action.AuthenticateOidcConfig = &elbv2.AuthenticateOidcActionConfig{
					AuthenticationRequestExtraParams: authenticationRequestExtraParams,
					AuthorizationEndpoint:            aws.String(authenticateOidcMap["authorization_endpoint"].(string)),
					ClientId:                         aws.String(authenticateOidcMap["client_id"].(string)),
					ClientSecret:                     aws.String(authenticateOidcMap["client_secret"].(string)),
					Issuer:                           aws.String(authenticateOidcMap["issuer"].(string)),
					TokenEndpoint:                    aws.String(authenticateOidcMap["token_endpoint"].(string)),
					UserInfoEndpoint:                 aws.String(authenticateOidcMap["user_info_endpoint"].(string)),
				}

				if onUnauthenticatedRequest, ok := authenticateOidcMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
					action.AuthenticateOidcConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
				}
				if scope, ok := authenticateOidcMap["scope"]; ok && scope != "" {
					action.AuthenticateOidcConfig.Scope = aws.String(scope.(string))
				}
				if sessionCookieName, ok := authenticateOidcMap["session_cookie_name"]; ok && sessionCookieName != "" {
					action.AuthenticateOidcConfig.SessionCookieName = aws.String(sessionCookieName.(string))
				}
				if sessionTimeout, ok := authenticateOidcMap["session_timeout"]; ok && sessionTimeout != 0 {
					action.AuthenticateOidcConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
				}
			} else {
				return errors.New("for actions of type 'authenticate-oidc', you must specify a 'authenticate_oidc' block")
			}
		}

		params.Actions[i] = action
	}

	var err error
	params.Conditions, err = lbListenerRuleConditions(d.Get("condition").(*schema.Set).List())
	if err != nil {
		return err
	}

	var resp *elbv2.CreateRuleOutput
	if v, ok := d.GetOk("priority"); ok {
		var err error
		params.Priority = aws.Int64(int64(v.(int)))
		resp, err = elbconn.CreateRule(params)
		if err != nil {
			return fmt.Errorf("Error creating LB Listener Rule: %v", err)
		}
	} else {
		var priority int64
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			var err error
			priority, err = highestListenerRulePriority(elbconn, listenerArn)
			if err != nil {
				return resource.NonRetryableError(err)
			}
			params.Priority = aws.Int64(priority + 1)
			resp, err = elbconn.CreateRule(params)
			if err != nil {
				if isAWSErr(err, elbv2.ErrCodePriorityInUseException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			priority, err = highestListenerRulePriority(elbconn, listenerArn)
			if err != nil {
				return fmt.Errorf("Error getting highest listener rule priority: %s", err)
			}
			params.Priority = aws.Int64(priority + 1)
			resp, err = elbconn.CreateRule(params)
		}
		if err != nil {
			return fmt.Errorf("Error creating LB Listener Rule: %v", err)
		}
	}

	if resp == nil || len(resp.Rules) == 0 {
		return errors.New("Error creating LB Listener Rule: no rules returned in response")
	}

	d.SetId(aws.StringValue(resp.Rules[0].RuleArn))

	return resourceAwsLbListenerRuleRead(d, meta)
}

func resourceAwsLbListenerRuleRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	var resp *elbv2.DescribeRulesOutput
	var req = &elbv2.DescribeRulesInput{
		RuleArns: []*string{aws.String(d.Id())},
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = elbconn.DescribeRules(req)
		if err != nil {
			if d.IsNewResource() && isAWSErr(err, elbv2.ErrCodeRuleNotFoundException, "") {
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(err)
			}
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		resp, err = elbconn.DescribeRules(req)
	}
	if err != nil {
		if isAWSErr(err, elbv2.ErrCodeRuleNotFoundException, "") {
			log.Printf("[WARN] DescribeRules - removing %s from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Rules for listener %q: %s", d.Id(), err)
	}

	if len(resp.Rules) != 1 {
		return fmt.Errorf("Error retrieving Rule %q", d.Id())
	}

	rule := resp.Rules[0]

	d.Set("arn", rule.RuleArn)

	// The listener arn isn't in the response but can be derived from the rule arn
	d.Set("listener_arn", lbListenerARNFromRuleARN(aws.StringValue(rule.RuleArn)))

	// Rules are evaluated in priority order, from the lowest value to the highest value. The default rule has the lowest priority.
	if aws.StringValue(rule.Priority) == "default" {
		d.Set("priority", 99999)
	} else {
		if priority, err := strconv.Atoi(aws.StringValue(rule.Priority)); err != nil {
			return fmt.Errorf("Cannot convert rule priority %q to int: %s", aws.StringValue(rule.Priority), err)
		} else {
			d.Set("priority", priority)
		}
	}

	sort.Slice(rule.Actions, func(i, j int) bool {
		return aws.Int64Value(rule.Actions[i].Order) < aws.Int64Value(rule.Actions[j].Order)
	})
	actions := make([]interface{}, len(rule.Actions))
	for i, action := range rule.Actions {
		actionMap := make(map[string]interface{})
		actionMap["type"] = aws.StringValue(action.Type)
		actionMap["order"] = aws.Int64Value(action.Order)

		switch actionMap["type"] {
		case elbv2.ActionTypeEnumForward:
			if aws.StringValue(action.TargetGroupArn) != "" {
				actionMap["target_group_arn"] = aws.StringValue(action.TargetGroupArn)
			} else {
				targetGroups := make([]map[string]interface{}, 0, len(action.ForwardConfig.TargetGroups))
				for _, targetGroup := range action.ForwardConfig.TargetGroups {
					targetGroups = append(targetGroups,
						map[string]interface{}{
							"arn":    aws.StringValue(targetGroup.TargetGroupArn),
							"weight": aws.Int64Value(targetGroup.Weight),
						},
					)
				}
				actionMap["forward"] = []map[string]interface{}{
					{
						"target_group": targetGroups,
						"stickiness": []map[string]interface{}{
							{
								"enabled":  aws.BoolValue(action.ForwardConfig.TargetGroupStickinessConfig.Enabled),
								"duration": aws.Int64Value(action.ForwardConfig.TargetGroupStickinessConfig.DurationSeconds),
							},
						},
					},
				}
			}

		case elbv2.ActionTypeEnumRedirect:
			actionMap["redirect"] = []map[string]interface{}{
				{
					"host":        aws.StringValue(action.RedirectConfig.Host),
					"path":        aws.StringValue(action.RedirectConfig.Path),
					"port":        aws.StringValue(action.RedirectConfig.Port),
					"protocol":    aws.StringValue(action.RedirectConfig.Protocol),
					"query":       aws.StringValue(action.RedirectConfig.Query),
					"status_code": aws.StringValue(action.RedirectConfig.StatusCode),
				},
			}

		case elbv2.ActionTypeEnumFixedResponse:
			actionMap["fixed_response"] = []map[string]interface{}{
				{
					"content_type": aws.StringValue(action.FixedResponseConfig.ContentType),
					"message_body": aws.StringValue(action.FixedResponseConfig.MessageBody),
					"status_code":  aws.StringValue(action.FixedResponseConfig.StatusCode),
				},
			}

		case elbv2.ActionTypeEnumAuthenticateCognito:
			authenticationRequestExtraParams := make(map[string]interface{})
			for key, value := range action.AuthenticateCognitoConfig.AuthenticationRequestExtraParams {
				authenticationRequestExtraParams[key] = aws.StringValue(value)
			}

			actionMap["authenticate_cognito"] = []map[string]interface{}{
				{
					"authentication_request_extra_params": authenticationRequestExtraParams,
					"on_unauthenticated_request":          aws.StringValue(action.AuthenticateCognitoConfig.OnUnauthenticatedRequest),
					"scope":                               aws.StringValue(action.AuthenticateCognitoConfig.Scope),
					"session_cookie_name":                 aws.StringValue(action.AuthenticateCognitoConfig.SessionCookieName),
					"session_timeout":                     aws.Int64Value(action.AuthenticateCognitoConfig.SessionTimeout),
					"user_pool_arn":                       aws.StringValue(action.AuthenticateCognitoConfig.UserPoolArn),
					"user_pool_client_id":                 aws.StringValue(action.AuthenticateCognitoConfig.UserPoolClientId),
					"user_pool_domain":                    aws.StringValue(action.AuthenticateCognitoConfig.UserPoolDomain),
				},
			}

		case elbv2.ActionTypeEnumAuthenticateOidc:
			authenticationRequestExtraParams := make(map[string]interface{})
			for key, value := range action.AuthenticateOidcConfig.AuthenticationRequestExtraParams {
				authenticationRequestExtraParams[key] = aws.StringValue(value)
			}

			// The LB API currently provides no way to read the ClientSecret
			// Instead we passthrough the configuration value into the state
			clientSecret := d.Get("action." + strconv.Itoa(i) + ".authenticate_oidc.0.client_secret").(string)

			actionMap["authenticate_oidc"] = []map[string]interface{}{
				{
					"authentication_request_extra_params": authenticationRequestExtraParams,
					"authorization_endpoint":              aws.StringValue(action.AuthenticateOidcConfig.AuthorizationEndpoint),
					"client_id":                           aws.StringValue(action.AuthenticateOidcConfig.ClientId),
					"client_secret":                       clientSecret,
					"issuer":                              aws.StringValue(action.AuthenticateOidcConfig.Issuer),
					"on_unauthenticated_request":          aws.StringValue(action.AuthenticateOidcConfig.OnUnauthenticatedRequest),
					"scope":                               aws.StringValue(action.AuthenticateOidcConfig.Scope),
					"session_cookie_name":                 aws.StringValue(action.AuthenticateOidcConfig.SessionCookieName),
					"session_timeout":                     aws.Int64Value(action.AuthenticateOidcConfig.SessionTimeout),
					"token_endpoint":                      aws.StringValue(action.AuthenticateOidcConfig.TokenEndpoint),
					"user_info_endpoint":                  aws.StringValue(action.AuthenticateOidcConfig.UserInfoEndpoint),
				},
			}
		}

		actions[i] = actionMap
	}
	d.Set("action", actions)

	conditions := make([]interface{}, len(rule.Conditions))
	for i, condition := range rule.Conditions {
		conditionMap := make(map[string]interface{})

		// Deprecated: remove in next major version of provider
		conditionMap["field"] = aws.StringValue(condition.Field)
		conditionMap["values"] = aws.StringValueSlice(condition.Values)

		switch conditionMap["field"] {
		case "host-header":
			conditionMap["host_header"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.HostHeaderConfig.Values),
				},
			}

		case "http-header":
			conditionMap["http_header"] = []interface{}{
				map[string]interface{}{
					"http_header_name": aws.StringValue(condition.HttpHeaderConfig.HttpHeaderName),
					"values":           flattenStringSet(condition.HttpHeaderConfig.Values),
				},
			}

		case "http-request-method":
			conditionMap["http_request_method"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.HttpRequestMethodConfig.Values),
				},
			}

		case "path-pattern":
			conditionMap["path_pattern"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.PathPatternConfig.Values),
				},
			}

		case "query-string":
			values := make([]interface{}, len(condition.QueryStringConfig.Values))
			for k, value := range condition.QueryStringConfig.Values {
				values[k] = map[string]interface{}{
					"key":   aws.StringValue(value.Key),
					"value": aws.StringValue(value.Value),
				}
			}
			conditionMap["query_string"] = schema.NewSet(lbListenerRuleConditionQueryStringHash, values)

		case "source-ip":
			conditionMap["source_ip"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.SourceIpConfig.Values),
				},
			}
		}

		conditions[i] = conditionMap
	}
	if err := d.Set("condition", conditions); err != nil {
		return fmt.Errorf("error setting condition: %s", err)
	}

	return nil
}

func resourceAwsLbListenerRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	if d.HasChange("priority") {
		params := &elbv2.SetRulePrioritiesInput{
			RulePriorities: []*elbv2.RulePriorityPair{
				{
					RuleArn:  aws.String(d.Id()),
					Priority: aws.Int64(int64(d.Get("priority").(int))),
				},
			},
		}

		_, err := elbconn.SetRulePriorities(params)
		if err != nil {
			return err
		}
	}

	requestUpdate := false
	params := &elbv2.ModifyRuleInput{
		RuleArn: aws.String(d.Id()),
	}

	if d.HasChange("action") {
		actions := d.Get("action").([]interface{})
		params.Actions = make([]*elbv2.Action, len(actions))
		for i, action := range actions {
			actionMap := action.(map[string]interface{})

			action := &elbv2.Action{
				Order: aws.Int64(int64(i + 1)),
				Type:  aws.String(actionMap["type"].(string)),
			}

			if order, ok := actionMap["order"]; ok && order.(int) != 0 {
				action.Order = aws.Int64(int64(order.(int)))
			}

			switch actionMap["type"].(string) {
			case elbv2.ActionTypeEnumForward:
				if err := lbListenerRuleActionForward(actionMap, action); err != nil {
					return err
				}

			case elbv2.ActionTypeEnumRedirect:
				redirectList := actionMap["redirect"].([]interface{})

				if len(redirectList) == 1 {
					redirectMap := redirectList[0].(map[string]interface{})

					action.RedirectConfig = &elbv2.RedirectActionConfig{
						Host:       aws.String(redirectMap["host"].(string)),
						Path:       aws.String(redirectMap["path"].(string)),
						Port:       aws.String(redirectMap["port"].(string)),
						Protocol:   aws.String(redirectMap["protocol"].(string)),
						Query:      aws.String(redirectMap["query"].(string)),
						StatusCode: aws.String(redirectMap["status_code"].(string)),
					}
				} else {
					return errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
				}

			case elbv2.ActionTypeEnumFixedResponse:
				fixedResponseList := actionMap["fixed_response"].([]interface{})

				if len(fixedResponseList) == 1 {
					fixedResponseMap := fixedResponseList[0].(map[string]interface{})

					action.FixedResponseConfig = &elbv2.FixedResponseActionConfig{
						ContentType: aws.String(fixedResponseMap["content_type"].(string)),
						MessageBody: aws.String(fixedResponseMap["message_body"].(string)),
						StatusCode:  aws.String(fixedResponseMap["status_code"].(string)),
					}
				} else {
					return errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
				}

			case elbv2.ActionTypeEnumAuthenticateCognito:
				authenticateCognitoList := actionMap["authenticate_cognito"].([]interface{})

				if len(authenticateCognitoList) == 1 {
					authenticateCognitoMap := authenticateCognitoList[0].(map[string]interface{})

					authenticationRequestExtraParams := make(map[string]*string)
					for key, value := range authenticateCognitoMap["authentication_request_extra_params"].(map[string]interface{}) {
						authenticationRequestExtraParams[key] = aws.String(value.(string))
					}

					action.AuthenticateCognitoConfig = &elbv2.AuthenticateCognitoActionConfig{
						AuthenticationRequestExtraParams: authenticationRequestExtraParams,
						UserPoolArn:                      aws.String(authenticateCognitoMap["user_pool_arn"].(string)),
						UserPoolClientId:                 aws.String(authenticateCognitoMap["user_pool_client_id"].(string)),
						UserPoolDomain:                   aws.String(authenticateCognitoMap["user_pool_domain"].(string)),
					}

					if onUnauthenticatedRequest, ok := authenticateCognitoMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
						action.AuthenticateCognitoConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
					}
					if scope, ok := authenticateCognitoMap["scope"]; ok && scope != "" {
						action.AuthenticateCognitoConfig.Scope = aws.String(scope.(string))
					}
					if sessionCookieName, ok := authenticateCognitoMap["session_cookie_name"]; ok && sessionCookieName != "" {
						action.AuthenticateCognitoConfig.SessionCookieName = aws.String(sessionCookieName.(string))
					}
					if sessionTimeout, ok := authenticateCognitoMap["session_timeout"]; ok && sessionTimeout != 0 {
						action.AuthenticateCognitoConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
					}
				} else {
					return errors.New("for actions of type 'authenticate-cognito', you must specify a 'authenticate_cognito' block")
				}

			case elbv2.ActionTypeEnumAuthenticateOidc:
				authenticateOidcList := actionMap["authenticate_oidc"].([]interface{})

				if len(authenticateOidcList) == 1 {
					authenticateOidcMap := authenticateOidcList[0].(map[string]interface{})

					authenticationRequestExtraParams := make(map[string]*string)
					for key, value := range authenticateOidcMap["authentication_request_extra_params"].(map[string]interface{}) {
						authenticationRequestExtraParams[key] = aws.String(value.(string))
					}

					action.AuthenticateOidcConfig = &elbv2.AuthenticateOidcActionConfig{
						AuthenticationRequestExtraParams: authenticationRequestExtraParams,
						AuthorizationEndpoint:            aws.String(authenticateOidcMap["authorization_endpoint"].(string)),
						ClientId:                         aws.String(authenticateOidcMap["client_id"].(string)),
						ClientSecret:                     aws.String(authenticateOidcMap["client_secret"].(string)),
						Issuer:                           aws.String(authenticateOidcMap["issuer"].(string)),
						TokenEndpoint:                    aws.String(authenticateOidcMap["token_endpoint"].(string)),
						UserInfoEndpoint:                 aws.String(authenticateOidcMap["user_info_endpoint"].(string)),
					}

					if onUnauthenticatedRequest, ok := authenticateOidcMap["on_unauthenticated_request"]; ok && onUnauthenticatedRequest != "" {
						action.AuthenticateOidcConfig.OnUnauthenticatedRequest = aws.String(onUnauthenticatedRequest.(string))
					}
					if scope, ok := authenticateOidcMap["scope"]; ok && scope != "" {
						action.AuthenticateOidcConfig.Scope = aws.String(scope.(string))
					}
					if sessionCookieName, ok := authenticateOidcMap["session_cookie_name"]; ok && sessionCookieName != "" {
						action.AuthenticateOidcConfig.SessionCookieName = aws.String(sessionCookieName.(string))
					}
					if sessionTimeout, ok := authenticateOidcMap["session_timeout"]; ok && sessionTimeout != 0 {
						action.AuthenticateOidcConfig.SessionTimeout = aws.Int64(int64(sessionTimeout.(int)))
					}
				} else {
					return errors.New("for actions of type 'authenticate-oidc', you must specify a 'authenticate_oidc' block")
				}
			}

			params.Actions[i] = action
		}
		requestUpdate = true
	}

	if d.HasChange("condition") {
		var err error
		params.Conditions, err = lbListenerRuleConditions(d.Get("condition").(*schema.Set).List())
		if err != nil {
			return err
		}
		requestUpdate = true
	}

	if requestUpdate {
		resp, err := elbconn.ModifyRule(params)
		if err != nil {
			return fmt.Errorf("Error modifying LB Listener Rule: %s", err)
		}

		if len(resp.Rules) == 0 {
			return errors.New("Error modifying creating LB Listener Rule: no rules returned in response")
		}
	}

	return resourceAwsLbListenerRuleRead(d, meta)
}

func resourceAwsLbListenerRuleDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	_, err := elbconn.DeleteRule(&elbv2.DeleteRuleInput{
		RuleArn: aws.String(d.Id()),
	})
	if err != nil && !isAWSErr(err, elbv2.ErrCodeRuleNotFoundException, "") {
		return fmt.Errorf("Error deleting LB Listener Rule: %s", err)
	}
	return nil
}

func validateAwsLbListenerRulePriority(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 1 || (value > 50000 && value != 99999) {
		errors = append(errors, fmt.Errorf("%q must be in the range 1-50000 for normal rule or 99999 for default rule", k))
	}
	return
}

// from arn:
// arn:aws:elasticloadbalancing:us-east-1:012345678912:listener-rule/app/name/0123456789abcdef/abcdef0123456789/456789abcedf1234
// select submatches:
// (arn:aws:elasticloadbalancing:us-east-1:012345678912:listener)-rule(/app/name/0123456789abcdef/abcdef0123456789)/456789abcedf1234
// concat to become:
// arn:aws:elasticloadbalancing:us-east-1:012345678912:listener/app/name/0123456789abcdef/abcdef0123456789
var lbListenerARNFromRuleARNRegexp = regexp.MustCompile(`^(arn:.+:listener)-rule(/.+)/[^/]+$`)

func lbListenerARNFromRuleARN(ruleArn string) string {
	if arnComponents := lbListenerARNFromRuleARNRegexp.FindStringSubmatch(ruleArn); len(arnComponents) > 1 {
		return arnComponents[1] + arnComponents[2]
	}

	return ""
}

func highestListenerRulePriority(conn *elbv2.ELBV2, arn string) (priority int64, err error) {
	var priorities []int
	var nextMarker *string

	for {
		out, aerr := conn.DescribeRules(&elbv2.DescribeRulesInput{
			ListenerArn: aws.String(arn),
			Marker:      nextMarker,
		})
		if aerr != nil {
			err = aerr
			return
		}
		for _, rule := range out.Rules {
			if aws.StringValue(rule.Priority) != "default" {
				p, _ := strconv.Atoi(aws.StringValue(rule.Priority))
				priorities = append(priorities, p)
			}
		}
		if out.NextMarker == nil {
			break
		}
		nextMarker = out.NextMarker
	}

	if len(priorities) == 0 {
		priority = 0
		return
	}

	sort.IntSlice(priorities).Sort()
	priority = int64(priorities[len(priorities)-1])

	return
}

// lbListenerRuleConditions converts data source generated by Terraform into
// an elbv2.RuleCondition object suitable for submitting to AWS API.
func lbListenerRuleConditions(conditions []interface{}) ([]*elbv2.RuleCondition, error) {
	elbConditions := make([]*elbv2.RuleCondition, len(conditions))
	for i, condition := range conditions {
		elbConditions[i] = &elbv2.RuleCondition{}
		conditionMap := condition.(map[string]interface{})
		var field string
		var attrs int

		if hostHeader, ok := conditionMap["host_header"].([]interface{}); ok && len(hostHeader) > 0 {
			field = "host-header"
			attrs += 1
			values := hostHeader[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].HostHeaderConfig = &elbv2.HostHeaderConditionConfig{
				Values: expandStringSet(values),
			}
		}

		if httpHeader, ok := conditionMap["http_header"].([]interface{}); ok && len(httpHeader) > 0 {
			field = "http-header"
			attrs += 1
			httpHeaderMap := httpHeader[0].(map[string]interface{})
			values := httpHeaderMap["values"].(*schema.Set)

			elbConditions[i].HttpHeaderConfig = &elbv2.HttpHeaderConditionConfig{
				HttpHeaderName: aws.String(httpHeaderMap["http_header_name"].(string)),
				Values:         expandStringSet(values),
			}
		}

		if httpRequestMethod, ok := conditionMap["http_request_method"].([]interface{}); ok && len(httpRequestMethod) > 0 {
			field = "http-request-method"
			attrs += 1
			values := httpRequestMethod[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].HttpRequestMethodConfig = &elbv2.HttpRequestMethodConditionConfig{
				Values: expandStringSet(values),
			}
		}

		if pathPattern, ok := conditionMap["path_pattern"].([]interface{}); ok && len(pathPattern) > 0 {
			field = "path-pattern"
			attrs += 1
			values := pathPattern[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].PathPatternConfig = &elbv2.PathPatternConditionConfig{
				Values: expandStringSet(values),
			}
		}

		if queryString, ok := conditionMap["query_string"].(*schema.Set); ok && queryString.Len() > 0 {
			field = "query-string"
			attrs += 1
			values := queryString.List()

			elbConditions[i].QueryStringConfig = &elbv2.QueryStringConditionConfig{
				Values: make([]*elbv2.QueryStringKeyValuePair, len(values)),
			}
			for j, p := range values {
				valuePair := p.(map[string]interface{})
				elbValuePair := &elbv2.QueryStringKeyValuePair{
					Value: aws.String(valuePair["value"].(string)),
				}
				if valuePair["key"].(string) != "" {
					elbValuePair.Key = aws.String(valuePair["key"].(string))
				}
				elbConditions[i].QueryStringConfig.Values[j] = elbValuePair
			}
		}

		if sourceIp, ok := conditionMap["source_ip"].([]interface{}); ok && len(sourceIp) > 0 {
			field = "source-ip"
			attrs += 1
			values := sourceIp[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].SourceIpConfig = &elbv2.SourceIpConditionConfig{
				Values: expandStringSet(values),
			}
		}

		// Deprecated backwards compatibility
		// This code is also hit during an update when the condition has not been modified. Issues: GH-11232 and GH-11362
		if cmField, ok := conditionMap["field"].(string); ok && (cmField == "host-header" || cmField == "path-pattern") {
			// When the condition is not being updated Terraform feeds in the existing state which has host header and
			// path pattern set in both locations with identical values.
			if field == cmField {
				values := schema.NewSet(schema.HashString, conditionMap["values"].([]interface{}))
				var values2 *schema.Set
				if cmField == "host-header" {
					values2 = conditionMap["host_header"].([]interface{})[0].(map[string]interface{})["values"].(*schema.Set)
				} else {
					values2 = conditionMap["path_pattern"].([]interface{})[0].(map[string]interface{})["values"].(*schema.Set)
				}
				if !values2.Equal(values) {
					attrs += 1
				}
			} else {
				field = cmField
				attrs += 1
				values := conditionMap["values"].([]interface{})
				if len(values) == 0 {
					return nil, errors.New("Both field and values must be set in a condition block")
				}
				elbConditions[i].Values = expandStringList(values)
			}
		}

		// FIXME Rework this and use ConflictsWith when it finally works with collections:
		// https://github.com/hashicorp/terraform/issues/13016
		// Still need to ensure that one of the condition attributes is set.
		if attrs == 0 {
			return nil, errors.New("One of host_header, http_header, http_request_method, path_pattern, query_string or source_ip must be set in a condition block")
		} else if attrs > 1 {
			// Deprecated: remove `field` from message
			return nil, errors.New("Only one of field, host_header, http_header, http_request_method, path_pattern, query_string or source_ip can be set in a condition block")
		}

		elbConditions[i].Field = aws.String(field)
	}
	return elbConditions, nil
}

func lbListenerRuleActionForward(actionMap map[string]interface{}, action *elbv2.Action) error {
	forwardList := actionMap["forward"].([]interface{})
	targetGroupArn := actionMap["target_group_arn"].(string)

	if targetGroupArn != "" {
		action.TargetGroupArn = aws.String(targetGroupArn)
	} else if len(forwardList) == 1 {
		forwardMap := forwardList[0].(map[string]interface{})
		targetGroupsInput := forwardMap["target_group"].(*schema.Set).List()
		weightedTargetGroups := make([]*elbv2.TargetGroupTuple, len(targetGroupsInput))

		for i, input := range targetGroupsInput {
			weightedTargetGroup := input.(map[string]interface{})
			weightedTargetGroups[i] = &elbv2.TargetGroupTuple{
				TargetGroupArn: aws.String(weightedTargetGroup["arn"].(string)),
				Weight:         aws.Int64(int64(weightedTargetGroup["weight"].(int))),
			}
		}

		action.ForwardConfig = &elbv2.ForwardActionConfig{
			TargetGroups: weightedTargetGroups,
		}

		stickinessInput := forwardMap["stickiness"].([]interface{})
		if len(stickinessInput) != 0 {
			stickyInputMap := stickinessInput[0].(map[string]interface{})
			action.ForwardConfig.TargetGroupStickinessConfig = &elbv2.TargetGroupStickinessConfig{
				Enabled:         aws.Bool(stickyInputMap["enabled"].(bool)),
				DurationSeconds: aws.Int64(int64(stickyInputMap["duration"].(int))),
			}
		}
	} else {
		return errors.New("for actions of type 'forward', you must specify a 'forward' block or 'target_group_arn'")
	}
	return nil
}
