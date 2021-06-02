package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2InstanceType_basic(t *testing.T) {
	resourceBasic := "data.aws_ec2_instance_type.basic"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ec2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceEc2InstanceTypeBasic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceBasic, "auto_recovery_supported", "true"),
					resource.TestCheckResourceAttr(resourceBasic, "bare_metal", "false"),
					resource.TestCheckResourceAttr(resourceBasic, "burstable_performance_supported", "false"),
					resource.TestCheckResourceAttr(resourceBasic, "current_generation", "true"),
					resource.TestCheckResourceAttr(resourceBasic, "dedicated_hosts_supported", "true"),
					resource.TestCheckResourceAttr(resourceBasic, "default_cores", "1"),
					resource.TestCheckResourceAttr(resourceBasic, "default_threads_per_core", "2"),
					resource.TestCheckResourceAttr(resourceBasic, "default_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceBasic, "ebs_encryption_support", "supported"),
					resource.TestCheckResourceAttr(resourceBasic, "ebs_nvme_support", "unsupported"),
					resource.TestCheckResourceAttr(resourceBasic, "ebs_optimized_support", "default"),
					resource.TestCheckResourceAttr(resourceBasic, "efa_supported", "false"),
					resource.TestCheckResourceAttr(resourceBasic, "ena_support", "required"),
					resource.TestCheckResourceAttr(resourceBasic, "free_tier_eligible", "false"),
					resource.TestCheckResourceAttr(resourceBasic, "hibernation_supported", "true"),
					resource.TestCheckResourceAttr(resourceBasic, "hypervisor", "nitro"),
					resource.TestCheckResourceAttr(resourceBasic, "instance_storage_supported", "false"),
					resource.TestCheckResourceAttr(resourceBasic, "instance_type", "m5.large"),
					resource.TestCheckResourceAttr(resourceBasic, "ipv6_supported", "true"),
					resource.TestCheckResourceAttr(resourceBasic, "maximum_ipv4_addresses_per_interface", "10"),
					resource.TestCheckResourceAttr(resourceBasic, "maximum_ipv6_addresses_per_interface", "10"),
					resource.TestCheckResourceAttr(resourceBasic, "maximum_network_interfaces", "3"),
					resource.TestCheckResourceAttr(resourceBasic, "memory_size", "8192"),
					resource.TestCheckResourceAttr(resourceBasic, "network_performance", "Up to 10 Gigabit"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_architectures.0", "x86_64"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_placement_strategies.#", "3"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_placement_strategies.0", "cluster"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_placement_strategies.1", "partition"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_placement_strategies.2", "spread"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_root_device_types.#", "1"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_root_device_types.0", "ebs"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_usages_classes.#", "2"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_usages_classes.0", "on-demand"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_usages_classes.1", "spot"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_virtualization_types.#", "1"),
					resource.TestCheckResourceAttr(resourceBasic, "supported_virtualization_types.0", "hvm"),
					resource.TestCheckResourceAttr(resourceBasic, "sustained_clock_speed", "3.1"),
					resource.TestCheckResourceAttr(resourceBasic, "valid_cores.#", "1"),
					resource.TestCheckResourceAttr(resourceBasic, "valid_cores.0", "1"),
					resource.TestCheckResourceAttr(resourceBasic, "valid_threads_per_core.#", "2"),
					resource.TestCheckResourceAttr(resourceBasic, "valid_threads_per_core.0", "1"),
					resource.TestCheckResourceAttr(resourceBasic, "valid_threads_per_core.1", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2InstanceType_metal(t *testing.T) {
	resourceMetal := "data.aws_ec2_instance_type.metal"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ec2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceEc2InstanceTypeMetal,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceMetal, "ebs_performance_baseline_bandwidth", "19000"),
					resource.TestCheckResourceAttr(resourceMetal, "ebs_performance_baseline_throughput", "2375"),
					resource.TestCheckResourceAttr(resourceMetal, "ebs_performance_baseline_iops", "80000"),
					resource.TestCheckResourceAttr(resourceMetal, "ebs_performance_maximum_bandwidth", "19000"),
					resource.TestCheckResourceAttr(resourceMetal, "ebs_performance_maximum_throughput", "2375"),
					resource.TestCheckResourceAttr(resourceMetal, "ebs_performance_maximum_iops", "80000"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.#", "1"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.0.count", "8"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.0.size", "7500"),
					resource.TestCheckResourceAttr(resourceMetal, "instance_disks.0.type", "ssd"),
					resource.TestCheckResourceAttr(resourceMetal, "total_instance_storage", "60000"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2InstanceType_gpu(t *testing.T) {
	resourceGpu := "data.aws_ec2_instance_type.gpu"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ec2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceEc2InstanceTypeGpu,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceGpu, "gpus.#", "1"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.count", "1"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.manufacturer", "NVIDIA"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.memory_size", "8192"),
					resource.TestCheckResourceAttr(resourceGpu, "gpus.0.name", "M60"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2InstanceType_fpga(t *testing.T) {
	resourceFpga := "data.aws_ec2_instance_type.fpga"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ec2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceEc2InstanceTypeFgpa,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.#", "1"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.count", "1"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.manufacturer", "Xilinx"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.memory_size", "65536"),
					resource.TestCheckResourceAttr(resourceFpga, "fpgas.0.name", "Virtex UltraScale (VU9P)"),
					resource.TestCheckResourceAttr(resourceFpga, "total_fpga_memory", "65536"),
				),
			},
		},
	})
}

const testAccDataSourceEc2InstanceTypeBasic = `
data "aws_ec2_instance_type" "basic" {
  instance_type = "m5.large"
}
`

const testAccDataSourceEc2InstanceTypeMetal = `
data "aws_ec2_instance_type" "metal" {
  instance_type = "i3en.metal"
}
`

const testAccDataSourceEc2InstanceTypeGpu = `
data "aws_ec2_instance_type" "gpu" {
  instance_type = "g3.4xlarge"
}
`

const testAccDataSourceEc2InstanceTypeFgpa = `
data "aws_ec2_instance_type" "fpga" {
  instance_type = "f1.2xlarge"
}
`
