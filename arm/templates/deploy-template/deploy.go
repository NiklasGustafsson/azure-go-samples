package main

import (
	"fmt"
	"os"
	
	"github.com/Azure/azure-sdk-for-go/arm"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest"
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/to"
)

func main() {

	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Println("usage: deploy [parameter-file-name [template-file-name]]")
		return
	}
	
	deploymentName := "simplelinux"
	groupName := "templatetests"
	groupLocation := "West US"
		
	arm,err := helpers.AuthenticateForARM()
	if err != nil {
		fmt.Printf("Failed to authenticate: '%s'\n", err.Error())
		return
	}
	
	arm.RequestInspector = helpers.WithInspection()
	arm.ResponseInspector = helpers.ByInspecting()
	
	_,err = createResourceGroup(groupName, groupLocation, arm)
	if err != nil {
		fmt.Printf("Failed to create resource group '%s': '%s'\n", groupName, err.Error())
		return
	}

	var parameterLink *string
	var parameters map[string]interface{}

	var templateLink *string
	
	if len(os.Args) >= 2 { 
		pl := os.Args[1]
		parameterLink = &pl
	}
	if len(os.Args) >= 3 { 
		tl := os.Args[2]
		templateLink = &tl
	}
	
	if parameterLink != nil {
		parameters,err = helpers.ReadMap(*parameterLink)
		if err != nil {
			fmt.Printf("Failed to read parameter file '%s': '%s'\n", *parameterLink, err.Error())
			return
		}
		if p,ok := parameters["parameters"]; ok {
			parameters = p.(map[string]interface{})
		}
	} else {
		parameters = map[string]interface{} {
			"adminUsername": makeStringParameterValue("tmpltest"),
			"adminPassword": makeStringParameterValue("<<PLEASE EDIT>>"),
			"dnsLabelPrefix": makeStringParameterValue("<<MUST BE UNIQUE>>"),
			"ubuntuOSVersion": makeStringParameterValue("14.04.2-LTS"),
		}
	}
	
	var deploymentProps resources.DeploymentProperties
	
	if templateLink != nil {

		template,err := helpers.ReadMap(*templateLink)
		if err != nil {
			fmt.Printf("Failed to read template file '%s': '%s'\n", *templateLink, err.Error())
			return
		}

		deploymentProps = resources.DeploymentProperties {
			Template: &template,
			Parameters: &parameters,
			Mode: resources.Incremental,
		}

	} else {

		deploymentProps = resources.DeploymentProperties {
			TemplateLink: &resources.TemplateLink { 
				URI: to.StringPtr("https://raw.githubusercontent.com/NiklasGustafsson/azure-go-samples/master/arm/templates/deploy-template/template01.json"),
				ContentVersion: to.StringPtr("1.0.0.0"),
			},
			Parameters: &parameters,
			Mode: resources.Incremental,
		}

	}

	deployment,err := arm.Deployments().CreateOrUpdate(groupName, deploymentName, resources.Deployment { Properties: &deploymentProps  })
	if err != nil {
		if aerr,ok := err.(autorest.Error); ok {
			fmt.Printf("Failed to create resource deployment details: '%s'\n", aerr.Message());
		} else {
			fmt.Printf("Failed to create resource deployment: '%s'\n", err.Error())		
		}
		return
	}
	
	fmt.Printf("Created resource deployment '%s'\n", *deployment.Name)	
}

func createResourceGroup(name, location string, arm arm.Client) (group resources.ResourceGroup, err error) {
	
	rgc := arm.ResourceGroups()
	
	params := resources.ResourceGroup{Name:&name,Location:&location}
	
	group,err = rgc.CreateOrUpdate(name, params)
	if err != nil {
		err = fmt.Errorf("Failed to create resource group '%s' in location '%s': '%s'\n", name, location, err.Error())
		return
	}
	
	fmt.Printf("Created resource group '%s'\n", *group.Name)
	
	return
}

func makeStringParameterValue(value string) map[string]interface{} {
	return map[string]interface{}{ "value": value }
}