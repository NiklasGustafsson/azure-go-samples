package main

import (
	"fmt"

	"github.com/Azure/azure-go-samples/helpers"

	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/azure-sdk-for-go/arm"
	"github.com/Azure/azure-sdk-for-go/arm/compute"
)

func main() {

	client, err := helpers.AuthenticateForARM()
	if err != nil {
		fmt.Printf("Failed to authenticate: '%s'\n", err.Error())
		return
	}

	client.RequestInspector = helpers.WithInspection()
	client.ResponseInspector = helpers.ByInspecting()

	vm1, err :=
		client.CreateSimpleVM(
			to.StringPtr("vmgroup02"),
			to.StringPtr("West US"),
			arm.VMParameters{
				User:     "user1",
				Password: "Catch22",
				ImageReference: compute.ImageReference{
					Publisher: to.StringPtr("MicrosoftWindowsServer"),
					Offer:     to.StringPtr("WindowsServer"),
					Sku:       to.StringPtr("2012-R2-Datacenter"),
					Version:   to.StringPtr("latest"),
				},
				AvailabilitySet: to.StringPtr("avset1"),
				StorageAccount:  to.StringPtr("vmgroup01accnta44c9ea1"),
			},
		)

	if err != nil {
		fmt.Printf("Failed to create virtual machine: %s\n", err.Error())
		return
	}

	fmt.Printf("Created vm '%s'\n", *vm1.Name)

	vm2, err :=
		client.CreateSimpleVM(
			to.StringPtr("vmgroup01"),
			to.StringPtr("West US"),
			arm.VMParameters{
				User:     "user1",
				Password: "Catch22",
				ImageReference: compute.ImageReference{
					Publisher: to.StringPtr("Canonical"),
					Offer:     to.StringPtr("UbuntuServer"),
					Sku:       to.StringPtr("14.04.2-LTS"),
					Version:   to.StringPtr("latest"),
				},
				AvailabilitySet: to.StringPtr("avset1"),
				StorageAccount:  to.StringPtr("vmgroup01accnta44c9ea1"),
			},
		)

	if err != nil {
		fmt.Printf("Failed to create virtual machine: %s\n", err.Error())
		return
	}

	fmt.Printf("Created vm '%s'\n", *vm2.Name)
}
