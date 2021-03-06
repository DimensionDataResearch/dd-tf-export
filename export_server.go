package main

import (
	"compute-api/compute"
	"fmt"
)

func makeServerResourceName(uniquenessKey int) string {
	return fmt.Sprintf("server%02d", uniquenessKey)
}

const configurationTemplateServer = `
resource "ddcloud_server" "%s" {
	name                    = "%s"
	description             = "%s"
	admin_password          = "password"

	memory_gb               = %d
	cpu_count               = %d

	networkdomain           = "%s"
	primary_adapter_vlan    = "%s"
	primary_adapter_ipv4    = "%s"

	dns_primary             = "8.8.8.8"
	dns_secondary           = "8.8.4.4"

	osimage_name            = "%s"%s%s
}
`

func (exporter *Exporter) exportServer(server compute.Server, networkDomainID string, primaryVLANResourceName string, uniquenessKey int) error {
	diskConfiguration, err := exporter.exportServerDisks(server)
	if err != nil {
		return err
	}

	tagConfiguration, err := exporter.exportServerTags(server)
	if err != nil {
		return err
	}

	osImage, err := exporter.APIClient.GetOSImage(server.SourceImageID)
	if err != nil {
		return err
	}
	if osImage == nil {
		return fmt.Errorf("Cannot find OS image '%s'.", server.SourceImageID)
	}

	configuration := fmt.Sprintf(configurationTemplateServer,
		makeServerResourceName(uniquenessKey),
		server.Name,
		server.Description,
		server.MemoryGB,
		server.CPU.Count,
		networkDomainID,
		primaryVLANResourceName,
		*server.Network.PrimaryAdapter.PrivateIPv4Address,
		osImage.Name,
		diskConfiguration,
		tagConfiguration,
	)
	fmt.Println(configuration)

	return nil
}

const configurationTemplateServerDisk = `
	disk {
		scsi_unit_id        = %d
		size_gb             = %d
		speed               = "%s"
	}`

func (exporter *Exporter) exportServerDisks(server compute.Server) (diskConfiguration string, err error) {
	for _, disk := range server.Disks {
		diskConfiguration += fmt.Sprintf(configurationTemplateServerDisk,
			disk.SCSIUnitID,
			disk.SizeGB,
			disk.Speed,
		)
	}

	if !isEmpty(diskConfiguration) {
		diskConfiguration = "\n" + diskConfiguration
	}

	return
}

const configurationTemplateServerTag = `
	tag {
		name                = "%s"
		value               = "%s"
	}`

func (exporter *Exporter) exportServerTags(server compute.Server) (tagConfiguration string, err error) {
	var tags *compute.TagDetails
	tags, err = exporter.APIClient.GetAssetTags(server.ID, compute.AssetTypeServer, nil)
	if err != nil {
		return
	}
	if len(tags.Items) == 0 {
		return
	}

	for _, tag := range tags.Items {
		tagConfiguration += fmt.Sprintf(configurationTemplateServerTag,
			tag.Name,
			tag.Value,
		)
	}

	if !isEmpty(tagConfiguration) {
		tagConfiguration = "\n" + tagConfiguration
	}

	return
}
