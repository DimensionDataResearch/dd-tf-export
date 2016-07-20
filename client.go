package main

import (
	"compute-api/compute"
	"fmt"
	"os"
)

const (
	// EnvComputeUser is the name of the DD_COMPUTE_USER environment variable.
	EnvComputeUser = "DD_COMPUTE_USER"

	// EnvComputePassword is the name of the DD_COMPUTE_PASSWORD environment variable.
	EnvComputePassword = "DD_COMPUTE_PASSWORD"
)

func createClient(options programOptions) (client *compute.Client, err error) {
	username := os.Getenv(EnvComputeUser)
	if isEmpty(username) {
		err = fmt.Errorf("The %s environment variable is not defined. Please set it to the CloudControl user name you wish to use.", EnvComputeUser)

		return
	}

	password := os.Getenv(EnvComputePassword)
	if isEmpty(password) {
		err = fmt.Errorf("The %s environment variable is not defined. Please set it to the CloudControl password you wish to use.", EnvComputePassword)

		return
	}

	client = compute.NewClient(options.Region, username, password)

	return
}
