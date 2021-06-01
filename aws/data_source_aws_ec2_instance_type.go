package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsEc2InstanceType() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2InstanceTypeRead,

		Schema: map[string]*schema.Schema{
			"auto_recovery_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"bare_metal": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"burstable_performance_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"current_generation": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"dedicated_hosts_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"default_cores": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"default_threads_per_core": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"default_vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ebs_encryption_support": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ebs_nvme_support": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ebs_optimized_support": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ebs_performance_baseline_bandwidth": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ebs_performance_baseline_throughput": {
				Type:     schema.TypeFloat,
				Computed: true,
			},

			"ebs_performance_baseline_iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ebs_performance_maximum_bandwidth": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"ebs_performance_maximum_throughput": {
				Type:     schema.TypeFloat,
				Computed: true,
			},

			"ebs_performance_maximum_iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"efa_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ena_support": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"fpgas": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"free_tier_eligible": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"gpus": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"hibernation_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"hypervisor": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"inference_accelerators": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"manufacturer": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"instance_disks": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"instance_storage_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ipv6_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"maximum_ipv4_addresses_per_interface": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"maximum_ipv6_addresses_per_interface": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"maximum_network_interfaces": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"memory_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"network_performance": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"supported_architectures": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"supported_placement_strategies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"supported_root_device_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"supported_usages_classes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"supported_virtualization_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"sustained_clock_speed": {
				Type:     schema.TypeFloat,
				Computed: true,
			},

			"total_fpga_memory": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"total_gpu_memory": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"total_instance_storage": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"valid_cores": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"valid_threads_per_core": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourceAwsEc2InstanceTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DescribeInstanceTypesInput{}

	instanceType := d.Get("instance_type").(string)
	params.InstanceTypes = []*string{aws.String(instanceType)}
	log.Printf("[DEBUG] Reading instances types: %s", params)

	resp, err := conn.DescribeInstanceTypes(params)
	if err != nil {
		return err
	}
	if len(resp.InstanceTypes) == 0 {
		return fmt.Errorf("no Instance Type found for %s", instanceType)
	}
	if len(resp.InstanceTypes) > 1 {
		return fmt.Errorf("multiple instance types found for type %s", instanceType)
	}
	v := resp.InstanceTypes[0]
	d.Set("auto_recovery_supported", v.AutoRecoverySupported)
	d.Set("bare_metal", v.BareMetal)
	d.Set("burstable_performance_supported", v.BurstablePerformanceSupported)
	d.Set("current_generation", v.CurrentGeneration)
	d.Set("dedicated_hosts_supported", v.DedicatedHostsSupported)
	d.Set("default_cores", v.VCpuInfo.DefaultCores)
	d.Set("default_threads_per_core", v.VCpuInfo.DefaultThreadsPerCore)
	d.Set("default_vcpus", v.VCpuInfo.DefaultVCpus)
	d.Set("ebs_encryption_support", v.EbsInfo.EncryptionSupport)
	d.Set("ebs_nvme_support", v.EbsInfo.NvmeSupport)
	d.Set("ebs_optimized_support", v.EbsInfo.EbsOptimizedSupport)
	if v.EbsInfo.EbsOptimizedInfo != nil {
		d.Set("ebs_performance_baseline_bandwidth", v.EbsInfo.EbsOptimizedInfo.BaselineBandwidthInMbps)
		d.Set("ebs_performance_baseline_throughput", v.EbsInfo.EbsOptimizedInfo.BaselineThroughputInMBps)
		d.Set("ebs_performance_baseline_iops", v.EbsInfo.EbsOptimizedInfo.BaselineIops)
		d.Set("ebs_performance_maximum_bandwidth", v.EbsInfo.EbsOptimizedInfo.MaximumBandwidthInMbps)
		d.Set("ebs_performance_maximum_throughput", v.EbsInfo.EbsOptimizedInfo.MaximumThroughputInMBps)
		d.Set("ebs_performance_maximum_iops", v.EbsInfo.EbsOptimizedInfo.MaximumIops)
	}
	d.Set("efa_supported", v.NetworkInfo.EfaSupported)
	d.Set("ena_support", v.NetworkInfo.EnaSupport)
	if v.FpgaInfo != nil {
		fpgaList := make([]interface{}, len(v.FpgaInfo.Fpgas))
		for i, fpg := range v.FpgaInfo.Fpgas {
			fpga := map[string]interface{}{
				"count":        aws.Int64Value(fpg.Count),
				"manufacturer": aws.StringValue(fpg.Manufacturer),
				"memory_size":  aws.Int64Value(fpg.MemoryInfo.SizeInMiB),
				"name":         aws.StringValue(fpg.Name),
			}
			fpgaList[i] = fpga
		}
		d.Set("fpgas", fpgaList)
		d.Set("total_fpga_memory", v.FpgaInfo.TotalFpgaMemoryInMiB)
	}
	d.Set("free_tier_eligible", v.FreeTierEligible)
	if v.GpuInfo != nil {
		gpuList := make([]interface{}, len(v.GpuInfo.Gpus))
		for i, gp := range v.GpuInfo.Gpus {
			gpu := map[string]interface{}{
				"count":        aws.Int64Value(gp.Count),
				"manufacturer": aws.StringValue(gp.Manufacturer),
				"memory_size":  aws.Int64Value(gp.MemoryInfo.SizeInMiB),
				"name":         aws.StringValue(gp.Name),
			}
			gpuList[i] = gpu
		}
		d.Set("gpus", gpuList)
		d.Set("total_gpu_memory", v.GpuInfo.TotalGpuMemoryInMiB)
	}
	d.Set("hibernation_supported", v.HibernationSupported)
	d.Set("hypervisor", v.Hypervisor)
	if v.InferenceAcceleratorInfo != nil {
		acceleratorList := make([]interface{}, len(v.InferenceAcceleratorInfo.Accelerators))
		for i, accl := range v.InferenceAcceleratorInfo.Accelerators {
			accelerator := map[string]interface{}{
				"count":        aws.Int64Value(accl.Count),
				"manufacturer": aws.StringValue(accl.Manufacturer),
				"name":         aws.StringValue(accl.Name),
			}
			acceleratorList[i] = accelerator
		}
		d.Set("inference_accelerators", acceleratorList)
	}
	if v.InstanceStorageInfo != nil {
		if v.InstanceStorageInfo.Disks != nil {
			diskList := make([]interface{}, len(v.InstanceStorageInfo.Disks))
			for i, dk := range v.InstanceStorageInfo.Disks {
				disk := map[string]interface{}{
					"count": aws.Int64Value(dk.Count),
					"size":  aws.Int64Value(dk.SizeInGB),
					"type":  aws.StringValue(dk.Type),
				}
				diskList[i] = disk
			}
			d.Set("instance_disks", diskList)
		}
		d.Set("total_instance_storage", v.InstanceStorageInfo.TotalSizeInGB)
	}
	d.Set("instance_storage_supported", v.InstanceStorageSupported)
	d.Set("instance_type", v.InstanceType)
	d.Set("ipv6_supported", v.NetworkInfo.Ipv6Supported)
	d.Set("maximum_ipv4_addresses_per_interface", v.NetworkInfo.Ipv4AddressesPerInterface)
	d.Set("maximum_ipv6_addresses_per_interface", v.NetworkInfo.Ipv6AddressesPerInterface)
	d.Set("maximum_network_interfaces", v.NetworkInfo.MaximumNetworkInterfaces)
	d.Set("memory_size", v.MemoryInfo.SizeInMiB)
	d.Set("network_performance", v.NetworkInfo.NetworkPerformance)
	d.Set("supported_architectures", v.ProcessorInfo.SupportedArchitectures)
	d.Set("supported_placement_strategies", v.PlacementGroupInfo.SupportedStrategies)
	d.Set("supported_root_device_types", v.SupportedRootDeviceTypes)
	d.Set("supported_usages_classes", v.SupportedUsageClasses)
	d.Set("supported_virtualization_types", v.SupportedVirtualizationTypes)
	d.Set("sustained_clock_speed", v.ProcessorInfo.SustainedClockSpeedInGhz)
	d.Set("valid_cores", v.VCpuInfo.ValidCores)
	d.Set("valid_threads_per_core", v.VCpuInfo.ValidThreadsPerCore)
	d.SetId(aws.StringValue(v.InstanceType))
	return nil
}
