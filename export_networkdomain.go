package main

import (
	"compute-api/compute"
	"fmt"
	"strings"
)

func makeNetworkDomainResourceName(uniquenessKey int) string {
	return fmt.Sprintf("networkdomain%02d", uniquenessKey)
}

const configurationTemplateNetworkDomain = `

/*
 * %s
 */

resource "ddcloud_networkdomain" "%s" {
	name                    = "%s"
	description             = "%s"
	datacenter              = "%s"
	plan                    = "%s"
}
`

// ExportNetworkDomain exports a ddcloud_networkdomain resource to Terraform configuration.
func (exporter *Exporter) ExportNetworkDomain(id string, uniquenessKey int, recurse bool) error {
	networkDomain, err := exporter.APIClient.GetNetworkDomain(id)
	if err != nil {
		return err
	}
	if networkDomain == nil {
		return fmt.Errorf("Cannot find network domain '%s'.", id)
	}

	networkDomainResourceName := makeNetworkDomainResourceName(uniquenessKey)
	networkDomainID := fmt.Sprintf("${ddcloud_networkdomain.%s.id}", networkDomainResourceName)

	configuration := strings.TrimSpace(
		fmt.Sprintf(configurationTemplateNetworkDomain,
			networkDomain.Name,
			networkDomainResourceName,
			networkDomain.Name,
			networkDomain.Description,
			networkDomain.DatacenterID,
			networkDomain.Type,
		),
	)
	fmt.Println(configuration)

	if !recurse {
		return nil
	}

	// Export VLANs
	fmt.Printf("\n/*\n * VLANs for %s\n */\n\n", networkDomain.Name)
	vlans, err := exporter.APIClient.ListVLANs(networkDomain.ID)
	if err != nil {
		return err
	}

	vlanResourceNamesByID := make(map[string]string)
	for index, vlan := range vlans.VLANs {
		vlanResourceName := makeVLANResourceName(uniquenessKey + index)
		vlanResourceNamesByID[vlan.ID] = vlanResourceName

		err = exporter.ExportVLAN(vlan, networkDomainID, uniquenessKey+index)
		if err != nil {
			return err
		}
	}

	// Export firewall rules
	fmt.Printf("\n/*\n * Firewall rules for %s\n */\n\n", networkDomain.Name)
	firewallRules, err := exporter.APIClient.ListFirewallRules(networkDomain.ID)
	for index, firewallRule := range firewallRules.Rules {
		err = exporter.ExportFirewallRule(firewallRule, networkDomainID, uniquenessKey+index)
		if err != nil {
			return err
		}
	}

	// Export servers
	fmt.Printf("\n/*\n * Servers for %s\n */\n\n", networkDomain.Name)
	serverResourceNamesByPrivateIPv4 := make(map[string]string)
	vlanResourceNamesByPrivateIPv4 := make(map[string]string)
	page := compute.DefaultPaging()
	for {
		servers, err := exporter.APIClient.ListServersInNetworkDomain(networkDomain.ID, page)
		if err != nil {
			return err
		}
		if servers.IsEmpty() {
			break // We're done.
		}

		for index, server := range servers.Items {
			vlanID := *server.Network.PrimaryAdapter.VLANID
			vlanResourceName, ok := vlanResourceNamesByID[vlanID]
			if ok {
				vlanID = fmt.Sprintf("${ddcloud_vlan.%s.id}", vlanResourceName)
			}

			serverPrivateIPv4 := *server.Network.PrimaryAdapter.PrivateIPv4Address
			vlanResourceNamesByPrivateIPv4[serverPrivateIPv4] = vlanResourceName

			serverUniquenessKey := uniquenessKey + index
			serverResourceName := makeServerResourceName(serverUniquenessKey)
			serverResourceNamesByPrivateIPv4[serverPrivateIPv4] = serverResourceName
			exporter.exportServer(server, networkDomainID, vlanID, serverUniquenessKey)
		}

		page.Next()
	}

	// Export NAT rules
	fmt.Printf("\n/*\n * NAT rules for %s\n */\n\n", networkDomain.Name)
	page = compute.DefaultPaging()
	for {
		natRules, err := exporter.APIClient.ListNATRules(networkDomain.ID, page)
		if err != nil {
			return err
		}
		if natRules.IsEmpty() {
			break // We're done.
		}

		for index, natRule := range natRules.Rules {
			vlanResourceName, ok := vlanResourceNamesByPrivateIPv4[natRule.InternalIPAddress]
			if !ok {
				vlanResourceName = ""
			}

			serverResourceName, ok := serverResourceNamesByPrivateIPv4[natRule.InternalIPAddress]
			if !ok {
				serverResourceName = ""
			}

			err = exporter.ExportNAT(natRule, networkDomainID, vlanResourceName, serverResourceName, uniquenessKey+index)
			if err != nil {
				return err
			}
		}

		page.Next()
	}

	return nil
}
