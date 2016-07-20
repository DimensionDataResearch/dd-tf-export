package main

import (
	"compute-api/compute"
	"fmt"
)

// Exporter is used to export CloudControl resources to Terraform configuration.
type Exporter struct {
	APIClient *compute.Client
}

const configurationTemplateProvider = `provider "ddcloud" {
	region = "%s"
}

`

// ExportProviderConfiguration exports the configuration for the ddcloud provider.
func (exporter *Exporter) ExportProviderConfiguration(region string) {
	fmt.Printf(configurationTemplateProvider, region)
}

func createExporter(options programOptions) (exporter *Exporter, err error) {
	var apiClient *compute.Client
	apiClient, err = createClient(options)
	if err != nil {
		return
	}

	exporter = &Exporter{
		APIClient: apiClient,
	}

	return
}
