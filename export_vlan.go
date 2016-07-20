package main

import (
	"compute-api/compute"
	"fmt"
	"strings"
)

func makeVLANResourceName(uniquenessKey int) string {
	return fmt.Sprintf("vlan%02d", uniquenessKey)
}

const configurationTemplateVLAN = `
resource "ddcloud_vlan" "%s" {
	name                    = "%s"
	description             = "%s"

	networkdomain           = "%s"

	ipv4_base_address       = "%s"
	ipv4_prefix_size        = %d
}
`

// ExportVLAN exports a ddcloud_vlan resource to Terraform configuration.
func (exporter *Exporter) ExportVLAN(vlan compute.VLAN, networkDomainID string, uniquenessKey int) error {
	configuration := strings.TrimSpace(
		fmt.Sprintf(configurationTemplateVLAN,
			makeVLANResourceName(uniquenessKey),
			vlan.Name,
			vlan.Description,
			networkDomainID,
			vlan.IPv4Range.BaseAddress,
			vlan.IPv4Range.PrefixSize,
		),
	)
	fmt.Println(configuration)

	return nil
}
