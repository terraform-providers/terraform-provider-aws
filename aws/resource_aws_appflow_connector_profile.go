package aws

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsAppFlowConnectorProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppFlowConnectorProfileCreate,
		Read:   resourceAwsAppFlowConnectorProfileRead,
		Update: resourceAwsAppFlowConnectorProfileUpdate,
		Delete: resourceAwsAppFlowConnectorProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"connection_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appflow.ConnectionMode_Values(), false),
			},
			"connector_profile_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connector_profile_config": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_profile_credentials": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"secret_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"datadog": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"application_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"dynatrace": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_token": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"google_analytics": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_secret": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
														},
													},
												},
												"refresh_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"honeycode": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
														},
													},
												},
												"refresh_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"infor_nexus": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_key_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"datakey": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"secret_access_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"user_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"marketo": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_secret": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
														},
													},
												},
											},
										},
									},
									"redshift": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_credentials_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
														},
													},
												},
												"refresh_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"service_now": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"singular": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"slack": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_secret": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
														},
													},
												},
											},
										},
									},
									"snowflake": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"trendmicro": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_secret_key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"veeva": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"password": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"username": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"zendesk": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"client_secret": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"oauth_request": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"auth_code": {
																Type:      schema.TypeString,
																Optional:  true,
																Sensitive: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"redirect_uri": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 512),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"connector_profile_properties": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"datadog": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"dynatrace": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"google_analytics": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"honeycode": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"infor_nexus": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"marketo": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"redshift": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"database_url": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"role_arn": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"is_sandbox_environment": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"service_now": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"singular": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"slack": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"snowflake": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"account_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"private_link_service_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"region": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 64),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"stage": {
													Type:     schema.TypeString,
													Required: true,
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														return old == new || old == "@"+new
													},
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"warehouse": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(0, 512),
														validation.StringMatch(regexp.MustCompile(`[\s\w/!@#+=.-]*`), "must match [\\s\\w/!@#+=.-]*"),
													),
												},
											},
										},
									},
									"trendmicro": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"veeva": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"zendesk": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_url": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 256),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"connector_profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexp.MustCompile(`[\w/!@#+=.-]+`), "must match [\\w/!@#+=.-]+"),
				),
			},
			"connector_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appflow.ConnectorType_Values(), false),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"last_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppFlowConnectorProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appflowconn
	name := d.Get("connector_profile_name").(string)

	createConnectorProfileInput := appflow.CreateConnectorProfileInput{
		ConnectionMode:         aws.String(d.Get("connection_mode").(string)),
		ConnectorProfileConfig: expandConnectorProfileConfig(d.Get("connector_profile_config").([]interface{})[0].(map[string]interface{})),
		ConnectorProfileName:   aws.String(name),
		ConnectorType:          aws.String(d.Get("connector_type").(string)),
	}

	if v, ok := d.Get("kms_arn").(string); ok && len(v) > 0 {
		createConnectorProfileInput.KmsArn = aws.String(v)
	}

	_, err := conn.CreateConnectorProfile(&createConnectorProfileInput)

	if err != nil {
		return fmt.Errorf("error creating AppFlow Connector Profile: %w", err)
	}

	d.SetId(name)

	return resourceAwsAppFlowConnectorProfileRead(d, meta)
}

func resourceAwsAppFlowConnectorProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appflowconn

	connectorProfile, err := getConnectorProfile(conn, d.Id())

	if err != nil {
		return err
	}

	credentials := d.Get("connector_profile_config.0.connector_profile_credentials").([]interface{})

	d.Set("connection_mode", connectorProfile.ConnectionMode)
	d.Set("connector_profile_arn", connectorProfile.ConnectorProfileArn)
	d.Set("connector_profile_name", connectorProfile.ConnectorProfileName)
	d.Set("connector_profile_config", flattenConnectorProfileConfig(connectorProfile.ConnectorProfileProperties, credentials))
	d.Set("connector_type", connectorProfile.ConnectorType)
	d.Set("created_at", aws.TimeValue(connectorProfile.CreatedAt).Format(time.RFC3339))
	d.Set("last_updated_at", aws.TimeValue(connectorProfile.LastUpdatedAt).Format(time.RFC3339))

	d.SetId(d.Get("connector_profile_name").(string))

	return nil
}

func getConnectorProfile(conn *appflow.Appflow, name string) (*appflow.ConnectorProfile, error) {
	params := &appflow.DescribeConnectorProfilesInput{
		ConnectorProfileNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeConnectorProfiles(params)

		if err != nil {
			return nil, err
		}

		for _, connectorProfile := range output.ConnectorProfileDetails {
			if aws.StringValue(connectorProfile.ConnectorProfileName) == name {
				return connectorProfile, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		params.NextToken = output.NextToken
	}

	return nil, fmt.Errorf("No AppFlow Connector Profile found with name: %s", name)
}

func resourceAwsAppFlowConnectorProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsAppFlowConnectorProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appflowconn
	_, err := conn.DeleteConnectorProfile(&appflow.DeleteConnectorProfileInput{
		ConnectorProfileName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error deleting AppFlow Connector Profile (%s): %w", d.Id(), err)
	}

	return nil
}

func expandConnectorProfileConfig(m map[string]interface{}) *appflow.ConnectorProfileConfig {
	cpc := appflow.ConnectorProfileConfig{
		ConnectorProfileCredentials: expandConnectorProfileCredentials(m["connector_profile_credentials"].([]interface{})[0].(map[string]interface{})),
		ConnectorProfileProperties:  expandConnectorProfileProperties(m["connector_profile_properties"].([]interface{})[0].(map[string]interface{})),
	}

	return &cpc
}

func expandConnectorProfileCredentials(m map[string]interface{}) *appflow.ConnectorProfileCredentials {
	cpc := appflow.ConnectorProfileCredentials{}

	if v, ok := m["amplitude"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Amplitude = expandAmplitudeConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["datadog"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Datadog = expandDatadogConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["dynatrace"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Dynatrace = expandDynatraceConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["google_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.GoogleAnalytics = expandGoogleAnalyticsConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["honeycode"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Honeycode = expandHoneycodeConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["infor_nexus"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.InforNexus = expandInforNexusConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["marketo"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Marketo = expandMarketoConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Salesforce = expandSalesforceConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["service_now"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.ServiceNow = expandServiceNowConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["singular"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Singular = expandSingularConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["slack"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Slack = expandSlackConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["snowflake"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Snowflake = expandSnowflakeConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["trendmicro"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Trendmicro = expandTrendmicroConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["veeva"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Veeva = expandVeevaConnectorProfileCredentials(v[0].(map[string]interface{}))
	}
	if v, ok := m["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Zendesk = expandZendeskConnectorProfileCredentials(v[0].(map[string]interface{}))
	}

	return &cpc
}

func expandAmplitudeConnectorProfileCredentials(m map[string]interface{}) *appflow.AmplitudeConnectorProfileCredentials {
	credentials := appflow.AmplitudeConnectorProfileCredentials{
		ApiKey:    aws.String(m["api_key"].(string)),
		SecretKey: aws.String(m["secret_key"].(string)),
	}

	return &credentials
}

func expandDatadogConnectorProfileCredentials(m map[string]interface{}) *appflow.DatadogConnectorProfileCredentials {
	credentials := appflow.DatadogConnectorProfileCredentials{
		ApiKey:         aws.String(m["api_key"].(string)),
		ApplicationKey: aws.String(m["application_key"].(string)),
	}

	return &credentials
}

func expandDynatraceConnectorProfileCredentials(m map[string]interface{}) *appflow.DynatraceConnectorProfileCredentials {
	credentials := appflow.DynatraceConnectorProfileCredentials{
		ApiToken: aws.String(m["api_token"].(string)),
	}

	return &credentials
}

func expandGoogleAnalyticsConnectorProfileCredentials(m map[string]interface{}) *appflow.GoogleAnalyticsConnectorProfileCredentials {
	credentials := appflow.GoogleAnalyticsConnectorProfileCredentials{
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandHoneycodeConnectorProfileCredentials(m map[string]interface{}) *appflow.HoneycodeConnectorProfileCredentials {
	credentials := appflow.HoneycodeConnectorProfileCredentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandInforNexusConnectorProfileCredentials(m map[string]interface{}) *appflow.InforNexusConnectorProfileCredentials {
	credentials := appflow.InforNexusConnectorProfileCredentials{
		AccessKeyId:     aws.String(m["access_key_id"].(string)),
		Datakey:         aws.String(m["datakey"].(string)),
		SecretAccessKey: aws.String(m["secret_access_key"].(string)),
		UserId:          aws.String(m["user_id"].(string)),
	}

	return &credentials
}

func expandMarketoConnectorProfileCredentials(m map[string]interface{}) *appflow.MarketoConnectorProfileCredentials {
	credentials := appflow.MarketoConnectorProfileCredentials{
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandSalesforceConnectorProfileCredentials(m map[string]interface{}) *appflow.SalesforceConnectorProfileCredentials {
	credentials := appflow.SalesforceConnectorProfileCredentials{}

	if v, ok := m["access_token"].(string); ok && v != "" {
		credentials.AccessToken = aws.String(v)
	}

	if v, ok := m["client_credentials_arn"].(string); ok && v != "" {
		credentials.ClientCredentialsArn = aws.String(v)
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	if v, ok := m["refresh_token"].(string); ok && v != "" {
		credentials.RefreshToken = aws.String(v)
	}

	return &credentials
}

func expandServiceNowConnectorProfileCredentials(m map[string]interface{}) *appflow.ServiceNowConnectorProfileCredentials {
	credentials := appflow.ServiceNowConnectorProfileCredentials{
		Password: aws.String(m["password"].(string)),
		Username: aws.String(m["username"].(string)),
	}

	return &credentials
}

func expandSingularConnectorProfileCredentials(m map[string]interface{}) *appflow.SingularConnectorProfileCredentials {
	credentials := appflow.SingularConnectorProfileCredentials{
		ApiKey: aws.String(m["api_key"].(string)),
	}

	return &credentials
}

func expandSlackConnectorProfileCredentials(m map[string]interface{}) *appflow.SlackConnectorProfileCredentials {
	credentials := appflow.SlackConnectorProfileCredentials{
		AccessToken:  aws.String(m["access_token"].(string)),
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandSnowflakeConnectorProfileCredentials(m map[string]interface{}) *appflow.SnowflakeConnectorProfileCredentials {
	credentials := appflow.SnowflakeConnectorProfileCredentials{
		Password: aws.String(m["password"].(string)),
		Username: aws.String(m["username"].(string)),
	}

	return &credentials
}

func expandTrendmicroConnectorProfileCredentials(m map[string]interface{}) *appflow.TrendmicroConnectorProfileCredentials {
	credentials := appflow.TrendmicroConnectorProfileCredentials{
		ApiSecretKey: aws.String(m["api_secret_key"].(string)),
	}

	return &credentials
}

func expandVeevaConnectorProfileCredentials(m map[string]interface{}) *appflow.VeevaConnectorProfileCredentials {
	credentials := appflow.VeevaConnectorProfileCredentials{
		Password: aws.String(m["password"].(string)),
		Username: aws.String(m["username"].(string)),
	}

	return &credentials
}

func expandZendeskConnectorProfileCredentials(m map[string]interface{}) *appflow.ZendeskConnectorProfileCredentials {
	credentials := appflow.ZendeskConnectorProfileCredentials{
		AccessToken:  aws.String(m["access_token"].(string)),
		ClientId:     aws.String(m["client_id"].(string)),
		ClientSecret: aws.String(m["client_secret"].(string)),
	}

	if v, ok := m["oauth_request"].([]interface{}); ok && len(v) > 0 {
		credentials.OAuthRequest = expandOAuthRequest(v[0].(map[string]interface{}))
	}

	return &credentials
}

func expandOAuthRequest(m map[string]interface{}) *appflow.ConnectorOAuthRequest {
	r := appflow.ConnectorOAuthRequest{}

	if v, ok := m["auth_code"].(string); ok && v != "" {
		r.AuthCode = aws.String(v)
	}

	if v, ok := m["redirect_uri"].(string); ok && v != "" {
		r.RedirectUri = aws.String(v)
	}

	return &r
}

func expandConnectorProfileProperties(m map[string]interface{}) *appflow.ConnectorProfileProperties {
	cpc := appflow.ConnectorProfileProperties{}

	if v, ok := m["amplitude"].([]interface{}); ok && len(v) > 0 {
		cpc.Amplitude = &appflow.AmplitudeConnectorProfileProperties{}
	}

	if v, ok := m["datadog"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Datadog = expandDatadogConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["dynatrace"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Dynatrace = expandDynatraceConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["google_analytics"].([]interface{}); ok && len(v) > 0 {
		cpc.GoogleAnalytics = &appflow.GoogleAnalyticsConnectorProfileProperties{}
	}

	if v, ok := m["honeycode"].([]interface{}); ok && len(v) > 0 {
		cpc.Honeycode = &appflow.HoneycodeConnectorProfileProperties{}
	}

	if v, ok := m["infor_nexus"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.InforNexus = expandInforNexusConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["marketo"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Marketo = expandMarketoConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["redshift"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Redshift = expandRedshiftConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Salesforce = expandSalesforceConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["service_now"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.ServiceNow = expandServiceNowConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["singular"].([]interface{}); ok && len(v) > 0 {
		cpc.Singular = &appflow.SingularConnectorProfileProperties{}
	}

	if v, ok := m["slack"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Slack = expandSlackConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["snowflake"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Snowflake = expandSnowflakeConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["trendmicro"].([]interface{}); ok && len(v) > 0 {
		cpc.Trendmicro = &appflow.TrendmicroConnectorProfileProperties{}
	}

	if v, ok := m["veeva"].([]interface{}); ok && len(v) > 0 {
		cpc.Veeva = expandVeevaConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	if v, ok := m["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cpc.Zendesk = expandZendeskConnectorProfileProperties(v[0].(map[string]interface{}))
	}

	return &cpc
}

func expandDatadogConnectorProfileProperties(m map[string]interface{}) *appflow.DatadogConnectorProfileProperties {
	properties := appflow.DatadogConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandDynatraceConnectorProfileProperties(m map[string]interface{}) *appflow.DynatraceConnectorProfileProperties {
	properties := appflow.DynatraceConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandInforNexusConnectorProfileProperties(m map[string]interface{}) *appflow.InforNexusConnectorProfileProperties {
	properties := appflow.InforNexusConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandMarketoConnectorProfileProperties(m map[string]interface{}) *appflow.MarketoConnectorProfileProperties {
	properties := appflow.MarketoConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandRedshiftConnectorProfileProperties(m map[string]interface{}) *appflow.RedshiftConnectorProfileProperties {
	properties := appflow.RedshiftConnectorProfileProperties{
		BucketName: aws.String(m["bucket_name"].(string)),
		RoleArn:    aws.String(m["role_arn"].(string)),
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		properties.BucketPrefix = aws.String(v)
	}

	if v, ok := m["database_url"].(string); ok && v != "" {
		properties.DatabaseUrl = aws.String(v)
	}

	return &properties
}

func expandServiceNowConnectorProfileProperties(m map[string]interface{}) *appflow.ServiceNowConnectorProfileProperties {
	properties := appflow.ServiceNowConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandSalesforceConnectorProfileProperties(m map[string]interface{}) *appflow.SalesforceConnectorProfileProperties {
	properties := appflow.SalesforceConnectorProfileProperties{}

	if v, ok := m["instance_url"].(string); ok && v != "" {
		properties.InstanceUrl = aws.String(v)
	}

	if v, ok := m["is_sandbox_environment"].(bool); ok {
		properties.IsSandboxEnvironment = aws.Bool(v)
	}

	return &properties
}

func expandSlackConnectorProfileProperties(m map[string]interface{}) *appflow.SlackConnectorProfileProperties {
	properties := appflow.SlackConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandSnowflakeConnectorProfileProperties(m map[string]interface{}) *appflow.SnowflakeConnectorProfileProperties {
	properties := appflow.SnowflakeConnectorProfileProperties{
		BucketName: aws.String(m["bucket_name"].(string)),
		Stage:      aws.String(m["stage"].(string)),
		Warehouse:  aws.String(m["warehouse"].(string)),
	}

	if v, ok := m["account_name"].(string); ok && v != "" {
		properties.AccountName = aws.String(v)
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		properties.BucketPrefix = aws.String(v)
	}

	if v, ok := m["private_link_service_name"].(string); ok && v != "" {
		properties.PrivateLinkServiceName = aws.String(v)
	}

	if v, ok := m["region"].(string); ok && v != "" {
		properties.Region = aws.String(v)
	}

	return &properties
}

func expandVeevaConnectorProfileProperties(m map[string]interface{}) *appflow.VeevaConnectorProfileProperties {
	properties := appflow.VeevaConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func expandZendeskConnectorProfileProperties(m map[string]interface{}) *appflow.ZendeskConnectorProfileProperties {
	properties := appflow.ZendeskConnectorProfileProperties{
		InstanceUrl: aws.String(m["instance_url"].(string)),
	}

	return &properties
}

func flattenConnectorProfileConfig(cpp *appflow.ConnectorProfileProperties, cpc []interface{}) []interface{} {
	m := make(map[string]interface{})

	m["connector_profile_properties"] = flattenConnectorProfileProperties(cpp)
	m["connector_profile_credentials"] = cpc

	return []interface{}{m}
}

func flattenConnectorProfileProperties(cpp *appflow.ConnectorProfileProperties) []interface{} {
	result := make(map[string]interface{})

	if cpp.Amplitude != nil {
		m := make(map[string]interface{})
		result["amplitude"] = []interface{}{m}
	}
	if cpp.Datadog != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.Datadog.InstanceUrl)
		result["datadog"] = []interface{}{m}
	}
	if cpp.Dynatrace != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.Dynatrace.InstanceUrl)
		result["dynatrace"] = []interface{}{m}
	}
	if cpp.GoogleAnalytics != nil {
		m := make(map[string]interface{})
		result["google_analytics"] = []interface{}{m}
	}
	if cpp.Honeycode != nil {
		m := make(map[string]interface{})
		result["honeycode"] = []interface{}{m}
	}
	if cpp.InforNexus != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.InforNexus.InstanceUrl)
		result["infor_nexus"] = []interface{}{m}
	}
	if cpp.Marketo != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.Marketo.InstanceUrl)
		result["marketo"] = []interface{}{m}
	}
	if cpp.Redshift != nil {
		result["redshift"] = flattenRedshiftConnectorProfileProperties(cpp.Redshift)
	}
	if cpp.Salesforce != nil {
		result["salesforce"] = flattenSalesforceConnectorProfileProperties(cpp.Salesforce)
	}
	if cpp.ServiceNow != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.ServiceNow.InstanceUrl)
		result["service_now"] = []interface{}{m}
	}
	if cpp.Singular != nil {
		m := make(map[string]interface{})
		result["singular"] = []interface{}{m}
	}
	if cpp.Slack != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.Slack.InstanceUrl)
		result["slack"] = []interface{}{m}
	}
	if cpp.Snowflake != nil {
		result["snowflake"] = flattenSnowflakeConnectorProfileProperties(cpp.Snowflake)
	}
	if cpp.Trendmicro != nil {
		m := make(map[string]interface{})
		result["trendmicro"] = []interface{}{m}
	}
	if cpp.Veeva != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.Veeva.InstanceUrl)
		result["veeva"] = []interface{}{m}
	}
	if cpp.Zendesk != nil {
		m := make(map[string]interface{})
		m["instance_url"] = aws.StringValue(cpp.Zendesk.InstanceUrl)
		result["zendesk"] = []interface{}{m}
	}

	return []interface{}{result}
}

func flattenRedshiftConnectorProfileProperties(properties *appflow.RedshiftConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	m["bucket_name"] = aws.StringValue(properties.BucketName)

	if properties.BucketPrefix != nil {
		m["bucket_prefix"] = aws.StringValue(properties.BucketPrefix)
	}

	if properties.DatabaseUrl != nil {
		m["database_url"] = aws.StringValue(properties.DatabaseUrl)
	}

	m["role_arn"] = aws.StringValue(properties.RoleArn)

	return []interface{}{m}
}

func flattenSalesforceConnectorProfileProperties(properties *appflow.SalesforceConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})

	if properties.InstanceUrl != nil {
		m["instance_url"] = aws.StringValue(properties.InstanceUrl)
	}

	if properties.IsSandboxEnvironment != nil {
		m["is_sandbox_environment"] = aws.BoolValue(properties.IsSandboxEnvironment)
	}

	return []interface{}{m}
}

func flattenSnowflakeConnectorProfileProperties(properties *appflow.SnowflakeConnectorProfileProperties) []interface{} {
	m := make(map[string]interface{})
	if properties.AccountName != nil {
		m["account_name"] = aws.StringValue(properties.AccountName)
	}

	m["bucket_name"] = aws.StringValue(properties.BucketName)

	if properties.BucketPrefix != nil {
		m["bucket_prefix"] = aws.StringValue(properties.BucketPrefix)
	}

	if properties.Region != nil {
		m["region"] = aws.StringValue(properties.Region)
	}

	m["stage"] = aws.StringValue(properties.Stage)
	m["warehouse"] = aws.StringValue(properties.Warehouse)

	return []interface{}{m}
}
