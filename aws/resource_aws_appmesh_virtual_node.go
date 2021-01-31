package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppmeshVirtualNode() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsAppmeshVirtualNodeCreate,
		Read:   resourceAwsAppmeshVirtualNodeRead,
		Update: resourceAwsAppmeshVirtualNodeUpdate,
		Delete: resourceAwsAppmeshVirtualNodeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshVirtualNodeImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsAppmeshVirtualNodeMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backend": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 0,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtual_service": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_service_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},

												"client_policy": appmeshVirtualNodeClientPolicySchema(),
											},
										},
									},
								},
							},
						},

						"backend_defaults": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_policy": appmeshVirtualNodeClientPolicySchema(),
								},
							},
						},

						"listener": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection_pool": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},

												"http": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_connections": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},

															"max_pending_requests": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},

												"http2": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},

												"tcp": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_connections": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},
											},
										},
									},

									"health_check": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"healthy_threshold": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(2, 10),
												},

												"interval_millis": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(5000, 300000),
												},

												"path": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"port": {
													Type:         schema.TypeInt,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IsPortNumber,
												},

												"protocol": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
												},

												"timeout_millis": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(2000, 60000),
												},

												"unhealthy_threshold": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(2, 10),
												},
											},
										},
									},

									"outlier_detection": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"base_ejection_duration": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},

												"interval": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},

												"max_ejection_percent": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 100),
												},

												"max_server_errors": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
											},
										},
									},

									"port_mapping": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},

												"protocol": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
												},
											},
										},
									},

									"timeout": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},

															"per_request": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},

												"http": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},

															"per_request": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},

												"http2": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},

															"per_request": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},

												"tcp": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"unit": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
																		},

																		"value": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},
											},
										},
									},

									"tls": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"certificate": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"acm": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"certificate_arn": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validateArn,
																		},
																	},
																},
																ExactlyOneOf: []string{"spec.0.listener.0.tls.0.certificate.0.acm", "spec.0.listener.0.tls.0.certificate.0.file"},
															},

															"file": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"certificate_chain": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"private_key": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
																ExactlyOneOf: []string{"spec.0.listener.0.tls.0.certificate.0.acm", "spec.0.listener.0.tls.0.certificate.0.file"},
															},
														},
													},
												},

												"mode": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.ListenerTlsMode_Values(), false),
												},
											},
										},
									},
								},
							},
						},

						"logging": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_log": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"file": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"path": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
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

						"service_discovery": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_cloud_map": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.service_discovery.0.dns"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"attributes": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												"namespace_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},

												"service_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},

									"dns": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.service_discovery.0.aws_cloud_map"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"hostname": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
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

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

// appmeshVirtualNodeClientPolicySchema returns the schema for `client_policy` attributes.
func appmeshVirtualNodeClientPolicySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tls": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enforce": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  true,
							},

							"ports": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeInt},
								Set:      schema.HashInt,
							},

							"validation": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"trust": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"acm": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"certificate_authority_arns": {
																	Type:     schema.TypeSet,
																	Required: true,
																	Elem:     &schema.Schema{Type: schema.TypeString},
																	Set:      schema.HashString,
																},
															},
														},
													},

													"file": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"certificate_chain": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 255),
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
						},
					},
				},
			},
		},
	}
}

func resourceAwsAppmeshVirtualNodeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	req := &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
		Spec:            expandAppmeshVirtualNodeSpec(d.Get("spec").([]interface{})),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppmeshTags(),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh virtual node: %s", req)
	resp, err := conn.CreateVirtualNode(req)

	if err != nil {
		return fmt.Errorf("error creating App Mesh virtual node: %w", err)
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))

	return resourceAwsAppmeshVirtualNodeRead(d, meta)
}

func resourceAwsAppmeshVirtualNodeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	resp, err := conn.DescribeVirtualNode(req)

	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading App Mesh virtual node (%s): %w", d.Id(), err)
	}

	if aws.StringValue(resp.VirtualNode.Status.Status) == appmesh.VirtualNodeStatusCodeDeleted {
		log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(resp.VirtualNode.Metadata.Arn)
	d.Set("name", resp.VirtualNode.VirtualNodeName)
	d.Set("mesh_name", resp.VirtualNode.MeshName)
	d.Set("mesh_owner", resp.VirtualNode.Metadata.MeshOwner)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualNode.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualNode.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("resource_owner", resp.VirtualNode.Metadata.ResourceOwner)
	err = d.Set("spec", flattenAppmeshVirtualNodeSpec(resp.VirtualNode.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %w", err)
	}

	tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh virtual node (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualNodeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateVirtualNodeInput{
			MeshName:        aws.String(d.Get("mesh_name").(string)),
			VirtualNodeName: aws.String(d.Get("name").(string)),
			Spec:            expandAppmeshVirtualNodeSpec(v.([]interface{})),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			req.MeshOwner = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating App Mesh virtual node: %s", req)
		_, err := conn.UpdateVirtualNode(req)

		if err != nil {
			return fmt.Errorf("error updating App Mesh virtual node (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh virtual node (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsAppmeshVirtualNodeRead(d, meta)
}

func resourceAwsAppmeshVirtualNodeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh virtual node: %s", d.Id())
	_, err := conn.DeleteVirtualNode(&appmesh.DeleteVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	})

	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting App Mesh virtual node (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsAppmeshVirtualNodeImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-node-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual node %s from mesh %s", name, mesh)

	conn := meta.(*AWSClient).appmeshconn

	resp, err := conn.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(mesh),
		VirtualNodeName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))
	d.Set("name", resp.VirtualNode.VirtualNodeName)
	d.Set("mesh_name", resp.VirtualNode.MeshName)

	return []*schema.ResourceData{d}, nil
}
