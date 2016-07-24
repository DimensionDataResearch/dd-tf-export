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

	page := compute.DefaultPaging()

	// Export VLANs
	vlanResourceNamesByID := make(map[string]string)
	fmt.Printf("\n/*\n * VLANs for %s\n */\n\n", networkDomain.Name)
	page.First()
	for {
		vlans, err := exporter.APIClient.ListVLANs(networkDomain.ID, page)
		if err != nil {
			return err
		}
		if vlans.IsEmpty() {
			break // We're done
		}

		for index, vlan := range vlans.VLANs {
			vlanResourceName := makeVLANResourceName(uniquenessKey + index)
			vlanResourceNamesByID[vlan.ID] = vlanResourceName

			err = exporter.ExportVLAN(vlan, networkDomainID, uniquenessKey+index)
			if err != nil {
				return err
			}
		}

		page.Next()
	}

	// Export servers
	fmt.Printf("\n/*\n * Servers for %s\n */\n\n", networkDomain.Name)
	serverResourceNamesByPrivateIPv4 := make(map[string]string)
	vlanResourceNamesByPrivateIPv4 := make(map[string]string)
	page.First()
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

			err = exporter.exportServer(server, networkDomainID, vlanID, serverUniquenessKey)
			if err != nil {
				return err
			}
		}

		page.Next()
	}

	// Export NAT rules
	fmt.Printf("\n/*\n * NAT rules for %s\n */\n\n", networkDomain.Name)
	natResourceNamesByPublicIPv4 := make(map[string]string)
	page.First()
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

			natResourceNamesByPublicIPv4[natRule.ExternalIPAddress] = makeNATResourceName(uniquenessKey + index)

			err = exporter.ExportNAT(natRule, networkDomainID, vlanResourceName, serverResourceName, uniquenessKey+index)
			if err != nil {
				return err
			}
		}

		page.Next()
	}

	// Export firewall rules
	fmt.Printf("\n/*\n * Firewall rules for %s\n */\n\n", networkDomain.Name)
	page.First()
	for {
		firewallRules, err := exporter.APIClient.ListFirewallRules(networkDomain.ID, page)
		if err != nil {
			return err
		}
		if firewallRules.IsEmpty() {
			break // We're done
		}

		for index, firewallRule := range firewallRules.Rules {
			var natResourceName string
			if firewallRule.Destination.IPAddress != nil {
				var ok bool
				natResourceName, ok = natResourceNamesByPublicIPv4[firewallRule.Destination.IPAddress.Address]

				if !ok {
					natResourceName = ""
				}
			}

			err = exporter.ExportFirewallRule(firewallRule, networkDomainID, natResourceName, uniquenessKey+index)
			if err != nil {
				return err
			}
		}

		page.Next()
	}

	return nil
}
