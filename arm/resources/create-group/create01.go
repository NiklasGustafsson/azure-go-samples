package main

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-go-samples/helpers"
)

func main() {
	
	groupName := "armtestgroup"
	groupLocation := "West US"

	spt,sid,err := helpers.AuthenticateForARM()
	if err != nil {
		fmt.Printf("Failed to authenticate: '%s'\n", err.Error())
		return
	}
	
	rgc := resources.NewGroupsClient(sid)
	rgc.Authorizer = spt	
	
	rgc.RequestInspector = helpers.WithInspection()
	rgc.ResponseInspector = helpers.ByInspecting()
	
	params := resources.ResourceGroup{Name:&groupName,Location:&groupLocation}
	
	group,err := rgc.CreateOrUpdate(groupName, params)
	if err != nil {
		fmt.Printf("Failed to create resource group '%s' in location '%s': '%s'\n", groupName, groupLocation, err.Error())
		return
	}
	
	fmt.Printf("Created resource group '%s'\n", *group.Name)
}