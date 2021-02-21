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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appmesh/finder"
)

func resourceAwsAppmeshVirtualGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshVirtualGatewayCreate,
		Read:   resourceAwsAppmeshVirtualGatewayRead,
		Update: resourceAwsAppmeshVirtualGatewayUpdate,
		Delete: resourceAwsAppmeshVirtualGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshVirtualGatewayImport,
		},

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
						"backend_defaults": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_policy": {
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
																Elem: &schema.Schema{
																	Type:         schema.TypeInt,
																	ValidateFunc: validation.IsPortNumber,
																},
																Set: schema.HashInt,
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
																									Elem: &schema.Schema{
																										Type:         schema.TypeString,
																										ValidateFunc: validateArn,
																									},
																									Set: schema.HashString,
																								},
																							},
																						},
																						ExactlyOneOf: []string{
																							"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm",
																							"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file",
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
																						ExactlyOneOf: []string{
																							"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm",
																							"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file",
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
							},
						},

						"listener": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
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
													ValidateFunc: validation.StringInSlice(appmesh.VirtualGatewayPortProtocol_Values(), false),
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
													ValidateFunc: validation.StringInSlice(appmesh.VirtualGatewayPortProtocol_Values(), false),
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
																ExactlyOneOf: []string{
																	"spec.0.listener.0.tls.0.certificate.0.acm",
																	"spec.0.listener.0.tls.0.certificate.0.file",
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

																		"private_key": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
																ExactlyOneOf: []string{
																	"spec.0.listener.0.tls.0.certificate.0.acm",
																	"spec.0.listener.0.tls.0.certificate.0.file",
																},
															},
														},
													},
												},

												"mode": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.VirtualGatewayListenerTlsMode_Values(), false),
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

func resourceAwsAppmeshVirtualGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	input := &appmesh.CreateVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		Spec:               expandAppmeshVirtualGatewaySpec(d.Get("spec").([]interface{})),
		Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppmeshTags(),
		VirtualGatewayName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh virtual gateway: %s", input)
	output, err := conn.CreateVirtualGateway(input)

	if err != nil {
		return fmt.Errorf("error creating App Mesh virtual gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.VirtualGateway.Metadata.Uid))

	return resourceAwsAppmeshVirtualGatewayRead(d, meta)
}

func resourceAwsAppmeshVirtualGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	virtualGateway, err := finder.VirtualGateway(conn, d.Get("mesh_name").(string), d.Get("name").(string), d.Get("mesh_owner").(string))

	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh virtual gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading App Mesh virtual gateway (%s): %w", d.Id(), err)
	}

	if virtualGateway == nil || aws.StringValue(virtualGateway.Status.Status) == appmesh.VirtualGatewayStatusCodeDeleted {
		log.Printf("[WARN] App Mesh virtual gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(virtualGateway.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", virtualGateway.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", virtualGateway.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", virtualGateway.MeshName)
	d.Set("mesh_owner", virtualGateway.Metadata.MeshOwner)
	d.Set("name", virtualGateway.VirtualGatewayName)
	d.Set("resource_owner", virtualGateway.Metadata.ResourceOwner)
	err = d.Set("spec", flattenAppmeshVirtualGatewaySpec(virtualGateway.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %w", err)
	}

	tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh virtual gateway (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		input := &appmesh.UpdateVirtualGatewayInput{
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			Spec:               expandAppmeshVirtualGatewaySpec(d.Get("spec").([]interface{})),
			VirtualGatewayName: aws.String(d.Get("name").(string)),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating App Mesh virtual gateway: %s", input)
		_, err := conn.UpdateVirtualGateway(input)

		if err != nil {
			return fmt.Errorf("error updating App Mesh virtual gateway (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh virtual gateway (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsAppmeshVirtualGatewayRead(d, meta)
}

func resourceAwsAppmeshVirtualGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh virtual gateway (%s)", d.Id())
	_, err := conn.DeleteVirtualGateway(&appmesh.DeleteVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get("name").(string)),
	})

	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting App Mesh virtual gateway (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsAppmeshVirtualGatewayImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-gateway-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual gateway %s from mesh %s", name, mesh)

	conn := meta.(*AWSClient).appmeshconn

	virtualGateway, err := finder.VirtualGateway(conn, mesh, name, "")

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(virtualGateway.Metadata.Uid))
	d.Set("mesh_name", virtualGateway.MeshName)
	d.Set("name", virtualGateway.VirtualGatewayName)

	return []*schema.ResourceData{d}, nil
}

func expandAppmeshVirtualGatewaySpec(vSpec []interface{}) *appmesh.VirtualGatewaySpec {
	if len(vSpec) == 0 || vSpec[0] == nil {
		return nil
	}

	spec := &appmesh.VirtualGatewaySpec{}

	mSpec := vSpec[0].(map[string]interface{})

	if vBackendDefaults, ok := mSpec["backend_defaults"].([]interface{}); ok && len(vBackendDefaults) > 0 && vBackendDefaults[0] != nil {
		backendDefaults := &appmesh.VirtualGatewayBackendDefaults{}

		mBackendDefaults := vBackendDefaults[0].(map[string]interface{})

		if vClientPolicy, ok := mBackendDefaults["client_policy"].([]interface{}); ok {
			backendDefaults.ClientPolicy = expandAppmeshVirtualGatewayClientPolicy(vClientPolicy)
		}

		spec.BackendDefaults = backendDefaults
	}

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []*appmesh.VirtualGatewayListener{}

		for _, vListener := range vListeners {
			listener := &appmesh.VirtualGatewayListener{}

			mListener := vListener.(map[string]interface{})

			if vHealthCheck, ok := mListener["health_check"].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				healthCheck := &appmesh.VirtualGatewayHealthCheckPolicy{}

				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int64(int64(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck["path"].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck["port"].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mHealthCheck["protocol"].(string); ok && vProtocol != "" {
					healthCheck.Protocol = aws.String(vProtocol)
				}
				if vTimeoutMillis, ok := mHealthCheck["timeout_millis"].(int); ok && vTimeoutMillis > 0 {
					healthCheck.TimeoutMillis = aws.Int64(int64(vTimeoutMillis))
				}
				if vUnhealthyThreshold, ok := mHealthCheck["unhealthy_threshold"].(int); ok && vUnhealthyThreshold > 0 {
					healthCheck.UnhealthyThreshold = aws.Int64(int64(vUnhealthyThreshold))
				}

				listener.HealthCheck = healthCheck
			}

			if vConnectionPool, ok := mListener["connection_pool"].([]interface{}); ok && len(vConnectionPool) > 0 && vConnectionPool[0] != nil {
				mConnectionPool := vConnectionPool[0].(map[string]interface{})

				connectionPool := &appmesh.VirtualGatewayConnectionPool{}

				if vGrpcConnectionPool, ok := mConnectionPool["grpc"].([]interface{}); ok && len(vGrpcConnectionPool) > 0 && vGrpcConnectionPool[0] != nil {
					mGrpcConnectionPool := vGrpcConnectionPool[0].(map[string]interface{})

					grpcConnectionPool := &appmesh.VirtualGatewayGrpcConnectionPool{}

					if vMaxRequests, ok := mGrpcConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						grpcConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Grpc = grpcConnectionPool
				}

				if vHttpConnectionPool, ok := mConnectionPool["http"].([]interface{}); ok && len(vHttpConnectionPool) > 0 && vHttpConnectionPool[0] != nil {
					mHttpConnectionPool := vHttpConnectionPool[0].(map[string]interface{})

					httpConnectionPool := &appmesh.VirtualGatewayHttpConnectionPool{}

					if vMaxConnections, ok := mHttpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						httpConnectionPool.MaxConnections = aws.Int64(int64(vMaxConnections))
					}
					if vMaxPendingRequests, ok := mHttpConnectionPool["max_pending_requests"].(int); ok && vMaxPendingRequests > 0 {
						httpConnectionPool.MaxPendingRequests = aws.Int64(int64(vMaxPendingRequests))
					}

					connectionPool.Http = httpConnectionPool
				}

				if vHttp2ConnectionPool, ok := mConnectionPool["http2"].([]interface{}); ok && len(vHttp2ConnectionPool) > 0 && vHttp2ConnectionPool[0] != nil {
					mHttp2ConnectionPool := vHttp2ConnectionPool[0].(map[string]interface{})

					http2ConnectionPool := &appmesh.VirtualGatewayHttp2ConnectionPool{}

					if vMaxRequests, ok := mHttp2ConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						http2ConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Http2 = http2ConnectionPool
				}

				listener.ConnectionPool = connectionPool
			}

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				portMapping := &appmesh.VirtualGatewayPortMapping{}

				mPortMapping := vPortMapping[0].(map[string]interface{})

				if vPort, ok := mPortMapping["port"].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mPortMapping["protocol"].(string); ok && vProtocol != "" {
					portMapping.Protocol = aws.String(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
				tls := &appmesh.VirtualGatewayListenerTls{}

				mTls := vTls[0].(map[string]interface{})

				if vMode, ok := mTls["mode"].(string); ok && vMode != "" {
					tls.Mode = aws.String(vMode)
				}

				if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					certificate := &appmesh.VirtualGatewayListenerTlsCertificate{}

					mCertificate := vCertificate[0].(map[string]interface{})

					if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
						acm := &appmesh.VirtualGatewayListenerTlsAcmCertificate{}

						mAcm := vAcm[0].(map[string]interface{})

						if vCertificateArn, ok := mAcm["certificate_arn"].(string); ok && vCertificateArn != "" {
							acm.CertificateArn = aws.String(vCertificateArn)
						}

						certificate.Acm = acm
					}

					if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
						file := &appmesh.VirtualGatewayListenerTlsFileCertificate{}

						mFile := vFile[0].(map[string]interface{})

						if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
							file.CertificateChain = aws.String(vCertificateChain)
						}
						if vPrivateKey, ok := mFile["private_key"].(string); ok && vPrivateKey != "" {
							file.PrivateKey = aws.String(vPrivateKey)
						}

						certificate.File = file
					}

					tls.Certificate = certificate
				}

				listener.Tls = tls
			}

			listeners = append(listeners, listener)
		}

		spec.Listeners = listeners
	}

	if vLogging, ok := mSpec["logging"].([]interface{}); ok && len(vLogging) > 0 && vLogging[0] != nil {
		logging := &appmesh.VirtualGatewayLogging{}

		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			accessLog := &appmesh.VirtualGatewayAccessLog{}

			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				file := &appmesh.VirtualGatewayFileAccessLog{}

				mFile := vFile[0].(map[string]interface{})

				if vPath, ok := mFile["path"].(string); ok && vPath != "" {
					file.Path = aws.String(vPath)
				}

				accessLog.File = file
			}

			logging.AccessLog = accessLog
		}

		spec.Logging = logging
	}

	return spec
}

func expandAppmeshVirtualGatewayClientPolicy(vClientPolicy []interface{}) *appmesh.VirtualGatewayClientPolicy {
	if len(vClientPolicy) == 0 || vClientPolicy[0] == nil {
		return nil
	}

	clientPolicy := &appmesh.VirtualGatewayClientPolicy{}

	mClientPolicy := vClientPolicy[0].(map[string]interface{})

	if vTls, ok := mClientPolicy["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
		tls := &appmesh.VirtualGatewayClientPolicyTls{}

		mTls := vTls[0].(map[string]interface{})

		if vEnforce, ok := mTls["enforce"].(bool); ok {
			tls.Enforce = aws.Bool(vEnforce)
		}

		if vPorts, ok := mTls["ports"].(*schema.Set); ok && vPorts.Len() > 0 {
			tls.Ports = expandInt64Set(vPorts)
		}

		if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
			validation := &appmesh.VirtualGatewayTlsValidationContext{}

			mValidation := vValidation[0].(map[string]interface{})

			if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
				trust := &appmesh.VirtualGatewayTlsValidationContextTrust{}

				mTrust := vTrust[0].(map[string]interface{})

				if vAcm, ok := mTrust["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
					acm := &appmesh.VirtualGatewayTlsValidationContextAcmTrust{}

					mAcm := vAcm[0].(map[string]interface{})

					if vCertificateAuthorityArns, ok := mAcm["certificate_authority_arns"].(*schema.Set); ok && vCertificateAuthorityArns.Len() > 0 {
						acm.CertificateAuthorityArns = expandStringSet(vCertificateAuthorityArns)
					}

					trust.Acm = acm
				}

				if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
					file := &appmesh.VirtualGatewayTlsValidationContextFileTrust{}

					mFile := vFile[0].(map[string]interface{})

					if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
						file.CertificateChain = aws.String(vCertificateChain)
					}

					trust.File = file
				}

				validation.Trust = trust
			}

			tls.Validation = validation
		}

		clientPolicy.Tls = tls
	}

	return clientPolicy
}

func flattenAppmeshVirtualGatewaySpec(spec *appmesh.VirtualGatewaySpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if backendDefaults := spec.BackendDefaults; backendDefaults != nil {
		mBackendDefaults := map[string]interface{}{
			"client_policy": flattenAppmeshVirtualGatewayClientPolicy(backendDefaults.ClientPolicy),
		}

		mSpec["backend_defaults"] = []interface{}{mBackendDefaults}
	}

	if spec.Listeners != nil && spec.Listeners[0] != nil {
		// Per schema definition, set at most 1 Listener
		listener := spec.Listeners[0]
		mListener := map[string]interface{}{}

		if connectionPool := listener.ConnectionPool; connectionPool != nil {
			mConnectionPool := map[string]interface{}{}

			if grpcConnectionPool := connectionPool.Grpc; grpcConnectionPool != nil {
				mGrpcConnectionPool := map[string]interface{}{
					"max_requests": int(aws.Int64Value(grpcConnectionPool.MaxRequests)),
				}
				mConnectionPool["grpc"] = []interface{}{mGrpcConnectionPool}
			}

			if httpConnectionPool := connectionPool.Http; httpConnectionPool != nil {
				mHttpConnectionPool := map[string]interface{}{
					"max_connections":      int(aws.Int64Value(httpConnectionPool.MaxConnections)),
					"max_pending_requests": int(aws.Int64Value(httpConnectionPool.MaxPendingRequests)),
				}
				mConnectionPool["http"] = []interface{}{mHttpConnectionPool}
			}

			if http2ConnectionPool := connectionPool.Http2; http2ConnectionPool != nil {
				mHttp2ConnectionPool := map[string]interface{}{
					"max_requests": int(aws.Int64Value(http2ConnectionPool.MaxRequests)),
				}
				mConnectionPool["http2"] = []interface{}{mHttp2ConnectionPool}
			}

			mListener["connection_pool"] = []interface{}{mConnectionPool}
		}

		if healthCheck := listener.HealthCheck; healthCheck != nil {
			mHealthCheck := map[string]interface{}{
				"healthy_threshold":   int(aws.Int64Value(healthCheck.HealthyThreshold)),
				"interval_millis":     int(aws.Int64Value(healthCheck.IntervalMillis)),
				"path":                aws.StringValue(healthCheck.Path),
				"port":                int(aws.Int64Value(healthCheck.Port)),
				"protocol":            aws.StringValue(healthCheck.Protocol),
				"timeout_millis":      int(aws.Int64Value(healthCheck.TimeoutMillis)),
				"unhealthy_threshold": int(aws.Int64Value(healthCheck.UnhealthyThreshold)),
			}
			mListener["health_check"] = []interface{}{mHealthCheck}
		}

		if portMapping := listener.PortMapping; portMapping != nil {
			mPortMapping := map[string]interface{}{
				"port":     int(aws.Int64Value(portMapping.Port)),
				"protocol": aws.StringValue(portMapping.Protocol),
			}
			mListener["port_mapping"] = []interface{}{mPortMapping}
		}

		if tls := listener.Tls; tls != nil {
			mTls := map[string]interface{}{
				"mode": aws.StringValue(tls.Mode),
			}

			if certificate := tls.Certificate; certificate != nil {
				mCertificate := map[string]interface{}{}

				if acm := certificate.Acm; acm != nil {
					mAcm := map[string]interface{}{
						"certificate_arn": aws.StringValue(acm.CertificateArn),
					}

					mCertificate["acm"] = []interface{}{mAcm}
				}

				if file := certificate.File; file != nil {
					mFile := map[string]interface{}{
						"certificate_chain": aws.StringValue(file.CertificateChain),
						"private_key":       aws.StringValue(file.PrivateKey),
					}

					mCertificate["file"] = []interface{}{mFile}
				}

				mTls["certificate"] = []interface{}{mCertificate}
			}

			mListener["tls"] = []interface{}{mTls}
		}

		mSpec["listener"] = []interface{}{mListener}
	}

	if logging := spec.Logging; logging != nil {
		mLogging := map[string]interface{}{}

		if accessLog := logging.AccessLog; accessLog != nil {
			mAccessLog := map[string]interface{}{}

			if file := accessLog.File; file != nil {
				mAccessLog["file"] = []interface{}{
					map[string]interface{}{
						"path": aws.StringValue(file.Path),
					},
				}
			}

			mLogging["access_log"] = []interface{}{mAccessLog}
		}

		mSpec["logging"] = []interface{}{mLogging}
	}

	return []interface{}{mSpec}
}

func flattenAppmeshVirtualGatewayClientPolicy(clientPolicy *appmesh.VirtualGatewayClientPolicy) []interface{} {
	if clientPolicy == nil {
		return []interface{}{}
	}

	mClientPolicy := map[string]interface{}{}

	if tls := clientPolicy.Tls; tls != nil {
		mTls := map[string]interface{}{
			"enforce": aws.BoolValue(tls.Enforce),
			"ports":   flattenInt64Set(tls.Ports),
		}

		if validation := tls.Validation; validation != nil {
			mValidation := map[string]interface{}{}

			if trust := validation.Trust; trust != nil {
				mTrust := map[string]interface{}{}

				if acm := trust.Acm; acm != nil {
					mAcm := map[string]interface{}{
						"certificate_authority_arns": flattenStringSet(acm.CertificateAuthorityArns),
					}

					mTrust["acm"] = []interface{}{mAcm}
				}

				if file := trust.File; file != nil {
					mFile := map[string]interface{}{
						"certificate_chain": aws.StringValue(file.CertificateChain),
					}

					mTrust["file"] = []interface{}{mFile}
				}

				mValidation["trust"] = []interface{}{mTrust}
			}

			mTls["validation"] = []interface{}{mValidation}
		}

		mClientPolicy["tls"] = []interface{}{mTls}
	}

	return []interface{}{mClientPolicy}
}
