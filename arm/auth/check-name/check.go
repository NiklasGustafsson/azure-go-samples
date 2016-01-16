package main

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/azure-sdk-for-go/arm"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-go-samples/helpers"
)


func main() {
	
	name := "storage-account-name"
	
	c, err := helpers.LoadCredentials()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	sid := c["subscriptionID"]
	tid := c["tenantID"]
	cid := c["clientID"]
	secret := c["clientSecret"]
	
	spt, err := azure.NewServicePrincipalToken(cid, secret, tid, azure.AzureResourceManagerScope)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	arm := arm.NewClient(sid, spt)
	arm.RequestInspector = helpers.WithInspection()
	arm.ResponseInspector = helpers.ByInspecting()

	ac := arm.StorageAccounts()

	cna, err := ac.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters {
			Name: to.StringPtr(name),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})

	if err != nil {
		log.Fatalf("Error: %v", err)
	} else {
		if to.Bool(cna.NameAvailable) {
			fmt.Printf("The name '%s' is available\n", name)
		} else {
			fmt.Printf("The name '%s' is unavailable because %s\n", name, to.String(cna.Message))
		}
	}
}