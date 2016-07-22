package main

import (
	"compute-api/compute"
	"fmt"
	"strings"
)

func makeNATResourceName(uniquenessKey int) string {
	return fmt.Sprintf("nat%02d", uniquenessKey)
}

const configurationTemplateNAT = `
resource "ddcloud_nat" "%s" {
	networkdomain           = "%s"
	private_ipv4            = "%s"

	# public_ipv4           = "%s"%s
}
`

const configurationTemplateNATDependsOnVLAN = `

	depends_on              = ["ddcloud_vlan.%s"]`

// ExportNAT exports a ddcloud_nat resource to Terraform configuration.
func (exporter *Exporter) ExportNAT(natRule compute.NATRule, networkDomainID string, vlanResourceName string, serverResourceName string, uniquenessKey int) error {
	natDependsOnVLANConfiguration := ""
	if vlanResourceName != "" {
		natDependsOnVLANConfiguration = fmt.Sprintf(configurationTemplateNATDependsOnVLAN, vlanResourceName)
	}

	natInternalAddress := natRule.InternalIPAddress
	if serverResourceName != "" {
		natInternalAddress = fmt.Sprintf("${ddcloud_server.%s.primary_adapter_ipv4}", serverResourceName)
	}

	configuration := strings.TrimSpace(
		fmt.Sprintf(configurationTemplateNAT,
			makeNATResourceName(uniquenessKey),
			networkDomainID,
			natInternalAddress,
			natRule.ExternalIPAddress,
			natDependsOnVLANConfiguration,
		),
	)
	fmt.Println(configuration)

	return nil
}
