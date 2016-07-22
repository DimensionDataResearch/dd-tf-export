package main

import (
	"fmt"
	"os"
)

func main() {
	options := parseOptions()

	if options.Version {
		fmt.Println("dd-tf-export " + ProgramVersion)

		os.Exit(0)
	}

	exporter, err := createExporter(options)
	if err != nil {
		fmt.Println(err)

		os.Exit(2)
	}

	exporter.ExportProviderConfiguration(options.Region)

	for index, networkDomainID := range options.NetworkDomains {
		err := exporter.ExportNetworkDomain(networkDomainID, index+1, !options.NoRecurse)
		if err != nil {
			fmt.Println(err)

			os.Exit(3)
		}
	}

	os.Exit(0)
}
