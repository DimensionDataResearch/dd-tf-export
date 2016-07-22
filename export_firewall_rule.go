package main

import (
	"compute-api/compute"
	"fmt"
	"strings"
)

func makeFirewallRuleResourceName(uniquenessKey int) string {
	return fmt.Sprintf("firewall_rule%02d", uniquenessKey)
}

const configurationTemplateFirewallRule = `
resource "ddcloud_firewall_rule" "%s" {
	name                    = "%s"
	placement               = "first"
	action                  = "%s"
	enabled                 = %t

	ip_version              = "%s"
	protocol                = "%s"

	%s= "%s"
	source_port             = "%s"

	%s= "%s"
	destination_port        = "%s"

	networkdomain           = "%s"
}
`

// ExportFirewallRule exports a ddcloud_firewallRule resource to Terraform configuration.
func (exporter *Exporter) ExportFirewallRule(firewallRule compute.FirewallRule, networkDomainID string, natResourceName string, uniquenessKey int) error {
	if firewallRule.RuleType == "DEFAULT_RULE" {
		return nil // Ignore built-in rules.
	}

	var (
		sourceLabel      string
		source           string
		sourcePort       string
		destinationLabel string
		destination      string
		destinationPort  string
	)
	if firewallRule.Source.Port != nil {
		sourcePort = fmt.Sprintf("%d", firewallRule.Source.Port.Begin)
	} else {
		sourcePort = "any"
	}

	sourceAddress := firewallRule.Source.IPAddress
	if sourceAddress != nil {
		if sourceAddress.PrefixSize != nil {
			sourceLabel = "source_network          "
			source = fmt.Sprintf("%s/%d", sourceAddress.Address, *sourceAddress.PrefixSize)
		} else {
			sourceLabel = "source_address          "
			source = strings.ToLower(firewallRule.Source.IPAddress.Address)
		}
	}
	if firewallRule.Destination.Port != nil {
		destinationPort = fmt.Sprintf("%d", firewallRule.Destination.Port.Begin)
	} else {
		destinationPort = "any"
	}
	destinationAddress := firewallRule.Destination.IPAddress
	if destinationAddress != nil {
		if !isEmpty(natResourceName) {
			destinationLabel = "destination_address     "
			destination = fmt.Sprintf("${ddcloud_nat.%s.public_ipv4}", natResourceName)
		} else if destinationAddress.PrefixSize != nil {
			destinationLabel = "destination_network     "
			destination = fmt.Sprintf("%s/%d", destinationAddress.Address, *destinationAddress.PrefixSize)
		} else {
			destinationLabel = "destination_address     "
			destination = strings.ToLower(firewallRule.Destination.IPAddress.Address)
		}
	}

	configuration := strings.TrimSpace(
		fmt.Sprintf(configurationTemplateFirewallRule,
			makeFirewallRuleResourceName(uniquenessKey),
			firewallRule.Name,
			convertFirewallAction(firewallRule.Action),
			firewallRule.Enabled,
			firewallRule.IPVersion,
			firewallRule.Protocol,
			sourceLabel,
			source,
			sourcePort,
			destinationLabel,
			destination,
			destinationPort,
			networkDomainID,
		),
	)
	fmt.Println(configuration)

	return nil
}

func convertFirewallAction(action string) string {
	switch action {
	case "ACCEPT_DECISIVELY":
		return "accept"
	case "DROP":
		return "DROP"
	default:
		return "UNKNOWN"
	}
}
