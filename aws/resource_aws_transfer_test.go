package aws

import (
	"testing"
)

func TestAccAWSTransfer_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Server": {
			"basic":                         testAccAWSTransferServer_basic,
			"disappears":                    testAccAWSTransferServer_disappears,
			"APIGateway":                    testAccAWSTransferServer_apiGateway,
			"APIGatewayForceDestroy":        testAccAWSTransferServer_apiGateway_forceDestroy,
			"Domain":                        testAccAWSTransferServer_domain,
			"ForceDestroy":                  testAccAWSTransferServer_forceDestroy,
			"HostKey":                       testAccAWSTransferServer_hostKey,
			"Protocols":                     testAccAWSTransferServer_protocols,
			"SecurityPolicy":                testAccAWSTransferServer_securityPolicy,
			"UpdateEndpointTypePublicToVPC": testAccAWSTransferServer_updateEndpointType_publicToVpc,
			"UpdateEndpointTypePublicToVPCAddressAllocationIDs":      testAccAWSTransferServer_updateEndpointType_publicToVpc_addressAllocationIds,
			"UpdateEndpointTypeVPCEndpointToVPC":                     testAccAWSTransferServer_updateEndpointType_vpcEndpointToVpc,
			"UpdateEndpointTypeVPCEndpointToVPCAddressAllocationIDs": testAccAWSTransferServer_updateEndpointType_vpcEndpointToVpc_addressAllocationIds,
			"UpdateEndpointTypeVPCEndpointToVPCSecurityGroupIDs":     testAccAWSTransferServer_updateEndpointType_vpcEndpointToVpc_securityGroupIds,
			"UpdateEndpointTypeVPCToPublic":                          testAccAWSTransferServer_updateEndpointType_vpcToPublic,
			"VPC":                                                    testAccAWSTransferServer_vpc,
			"VPCAddressAllocationIDs":                                testAccAWSTransferServer_vpcAddressAllocationIds,
			"VPCAddressAllocationIDsSecurityGroupIDs":                testAccAWSTransferServer_vpcAddressAllocationIds_securityGroupIds,
			"VPCEndpointID":                                          testAccAWSTransferServer_vpcEndpointId,
			"VPCSecurityGroupIDs":                                    testAccAWSTransferServer_vpcSecurityGroupIds,
		},
		"SSHKey": {
			"basic": testAccAWSTransferSshKey_basic,
		},
		"User": {
			"basic":                 testAccAWSTransferUser_basic,
			"disappears":            testAccAWSTransferUser_disappears,
			"HomeDirectoryMappings": testAccAWSTransferUser_homeDirectoryMappings,
			"ModifyWithOptions":     testAccAWSTransferUser_modifyWithOptions,
			"Posix":                 testAccAWSTransferUser_posix,
			"UserNameValidation":    testAccAWSTransferUser_UserName_Validation,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
