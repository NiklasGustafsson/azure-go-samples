package main

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest"
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/to"
)

func main() {
	
	groupName := "createvm01"
	groupLocation := "West US"

	var group resources.ResourceGroup
	var subnet network.Subnet
	var nic network.Interface
	var avset compute.AvailabilitySet
	
	spt,sid,err := helpers.AuthenticateForARM()
	if err != nil {
		fmt.Printf("Failed to authenticate: '%s'\n", err.Error())
		return
	}
	
	group,err = createResourceGroup(sid, groupName, groupLocation, spt)
	
	if err != nil {
		fmt.Printf("ERROR:'%s'\n", err.Error())
		return
	}
	
	if err := createStorageAccount(sid, group, spt); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return
	}

	if avset,err = createAvailabilitySet(sid, group, spt); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return	
	}
	
	if subnet,err = createNetwork(sid, group, spt); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return	
	}
	
	if nic,err = createNetworkInterface(sid, "01", group, subnet, spt); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return	
	}

	if err := createVirtualMachine(sid, group, "vm001", "admin", "foobar1234", avset, nic, spt); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return	
	}
}

func createResourceGroup(
	subscription, name, location string, 
	authorizer autorest.Authorizer) (group resources.ResourceGroup, err error) {
	
	rgc := resources.NewGroupsClient(subscription)
	rgc.Authorizer = authorizer	
	
	rgc.RequestInspector = helpers.WithInspection()
	rgc.ResponseInspector = helpers.ByInspecting()
	
	params := resources.ResourceGroup{Name:&name,Location:&location}
	
	group,err = rgc.CreateOrUpdate(name, params)
	if err != nil {
		err = fmt.Errorf("Failed to create resource group '%s' in location '%s': '%s'\n", name, location, err.Error())
		return
	}
	
	fmt.Printf("Created resource group '%s'\n", *group.Name)
	rpc := resources.NewProvidersClient(subscription)
	rpc.Authorizer = authorizer
	
	rpc.RequestInspector = helpers.WithInspection()
	rpc.ResponseInspector = helpers.ByInspecting()

	if _,err1 := rpc.Register("Microsoft.Storage"); err != nil {
		err = fmt.Errorf("Failed to register resource provider 'Microsoft.Storage': '%s'\n", err1.Error())
	}
	if _,err1 := rpc.Register("Microsoft.Network"); err != nil {
		err = fmt.Errorf("Failed to register resource provider 'Microsoft.Network': '%s'\n", err1.Error())
	}
	if _,err1 := rpc.Register("Microsoft.Compute"); err != nil {
		err = fmt.Errorf("Failed to register resource provider 'Microsoft.Compute': '%s'\n", err1.Error())
	}
	
	return
}

func createStorageAccount(
	subscription string, 
	group resources.ResourceGroup, 
	authorizer autorest.Authorizer) error {
	
	ac := storage.NewAccountsClient(subscription)
	ac.Authorizer = authorizer

	ac.RequestInspector = helpers.WithInspection()
	ac.ResponseInspector = helpers.ByInspecting()

	cna, err := ac.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters {
			Name: group.Name,
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})

	if err != nil {
		return err
	}
	
	if to.Bool(cna.NameAvailable) {
		
		name := *group.Name
		props := storage.AccountPropertiesCreateParameters {AccountType: storage.StandardLRS}
		
		_,err = ac.Create(name, name,
			storage.AccountCreateParameters {
				Name: group.Name,
				Location: group.Location,
				Properties: &props,
			})
				
		if err != nil {
			return fmt.Errorf("Failed to create storage account '%s' in location '%s': '%s'\n", name, *group.Location, err.Error())
		}
	}
	
	return nil
}

func createAvailabilitySet(
	subscription string, 
	group resources.ResourceGroup, 
	authorizer autorest.Authorizer) (result compute.AvailabilitySet, err error) {
	
	avsc := compute.NewAvailabilitySetsClient(subscription)
	
	avsc.Authorizer = authorizer; 
	avsc.RequestInspector = helpers.WithInspection() 
	avsc.ResponseInspector = helpers.ByInspecting()
	
	name := *group.Name

	result,err = avsc.CreateOrUpdate(name, name+"avset", compute.AvailabilitySet { Location: group.Location })
	if err != nil {
		err = fmt.Errorf("Failed to create availability set '%s' in location '%s': '%s'\n", name, *group.Location, err.Error())
		return
	}
	
	return result,nil
}

func createNetwork(
	subscription string, 
	group resources.ResourceGroup, 
	authorizer autorest.Authorizer) (snetResult network.Subnet, err error) {
	
	vnetc := network.NewVirtualNetworksClient(subscription)
	snetc := network.NewSubnetsClient(subscription)
	
	vnetc.Authorizer = authorizer; 
	snetc.Authorizer = authorizer; 

	vnetc.RequestInspector = helpers.WithInspection() 
	vnetc.ResponseInspector = helpers.ByInspecting()
	snetc.RequestInspector = helpers.WithInspection() 
	snetc.ResponseInspector = helpers.ByInspecting()

	name := *group.Name
	vnet := name+"vnet"
	subnet := name+"subnet"

	snet := network.Subnet {
		Name: &subnet, 
		Properties: &network.SubnetPropertiesFormat { AddressPrefix: to.StringPtr("10.0.0.0/24") } }
	snets := make([]network.Subnet, 1, 1)
	snets[0] = snet
	
	addrPrefixes := make([]string,1,1)
	addrPrefixes[0] = "10.0.0.0/16"
	address := network.AddressSpace { AddressPrefixes: &addrPrefixes }
	
	nwkProps := network.VirtualNetworkPropertiesFormat { AddressSpace: &address, Subnets: &snets }
	
	_,err = vnetc.CreateOrUpdate(name, vnet, network.VirtualNetwork { Location: group.Location, Properties: &nwkProps })
	if err != nil {
		err = fmt.Errorf("Failed to create virtual network '%s' in location '%s': '%s'\n", vnet, *group.Location, err.Error())
		return
	}
	
	snetResult, err = snetc.CreateOrUpdate(name, vnet, subnet, snet)
	if err != nil {
		err = fmt.Errorf("Failed to create subnet '%s' in location '%s': '%s'\n", subnet, *group.Location, err.Error())
	}
	
	return
}

func createNetworkInterface(
	subscription, suffix string, 
	group resources.ResourceGroup, 
	subnet network.Subnet, 
	authorizer autorest.Authorizer) (networkInterface network.Interface, err error) {
		
	pipc  := network.NewPublicIPAddressesClient(subscription)
	nicc  := network.NewInterfacesClient(subscription)

	pipc.Authorizer = authorizer; 
	nicc.Authorizer = authorizer; 

	pipc.RequestInspector = helpers.WithInspection() 
	pipc.ResponseInspector = helpers.ByInspecting()
	nicc.RequestInspector = helpers.WithInspection() 
	nicc.ResponseInspector = helpers.ByInspecting()

	groupName := *group.Name
	ipName := "ip"+suffix
	nicName := "nic"+suffix
	
	pipResult,err := pipc.CreateOrUpdate(
		groupName, 
		ipName, 
		network.PublicIPAddress {
			Location: group.Location, 
			Properties: &network.PublicIPAddressPropertiesFormat {
				PublicIPAllocationMethod: network.Dynamic,
			},
		})
		
	if err != nil {
		err = fmt.Errorf("Failed to create public ip address '%s' in location '%s': '%s'\n", ipName, *group.Location, err.Error())
		return
	}
	
	nicProps := network.InterfaceIPConfigurationPropertiesFormat{  
		PublicIPAddress: &network.SubResource { ID: pipResult.ID }, 
		Subnet: &network.SubResource { ID: subnet.ID } }
		
	ipConfigs := make([]network.InterfaceIPConfiguration,1,1)
	ipConfigs[0] = network.InterfaceIPConfiguration {
		Name: to.StringPtr(nicName+"Config"),
		Properties: &nicProps,
	}
	props := network.InterfacePropertiesFormat { IPConfigurations: &ipConfigs }
	
	networkInterface,err = nicc.CreateOrUpdate(
		groupName, 
		nicName, 
		network.Interface {
			Location: group.Location, 
			Properties: &props,
		})	
	if err != nil {
		err = fmt.Errorf("Failed to create network interface '%s' in location '%s': '%s'\n", nicName, *group.Location, err.Error())
	}
	
	return
}

func createVirtualMachine(
	subscription string, 
	group resources.ResourceGroup,
	vmName, adminName, adminPassword string, 
	availSet compute.AvailabilitySet, 
	networkInterface network.Interface, 
	authorizer autorest.Authorizer) error {
	
	vmc := compute.NewVirtualMachinesClient(subscription)

	vmc.Authorizer = authorizer	
	
	vmc.RequestInspector = helpers.WithInspection() 
	vmc.ResponseInspector = helpers.ByInspecting()

	netRefs := make([]compute.NetworkInterfaceReference,1,1)
	netRefs[0] = compute.NetworkInterfaceReference { ID: networkInterface.ID } 
	
	groupName := *group.Name
	accountName := groupName
	
	vmParams := compute.VirtualMachine { 
		Location: group.Location,
		Properties: &compute.VirtualMachineProperties { 
			AvailabilitySet: &compute.SubResource { ID: availSet.ID },
			HardwareProfile: &compute.HardwareProfile {VMSize: compute.StandardA0 },
			NetworkProfile: &compute.NetworkProfile { NetworkInterfaces: &netRefs },
			StorageProfile: &compute.StorageProfile { 
				ImageReference: &compute.ImageReference {
					Publisher: to.StringPtr("MicrosoftWindowsServer"),
					Offer: to.StringPtr("WindowsServer"),
					Sku: to.StringPtr("2012-R2-Datacenter"),
					Version: to.StringPtr("latest"),
				}, 
				OsDisk: &compute.OSDisk{
					Name: to.StringPtr("mytestod1"),
					CreateOption: compute.FromImage,
					Vhd: &compute.VirtualHardDisk { 
						URI: to.StringPtr("http://" + accountName + ".blob.core.windows.net/vhds/mytestod1.vhd"),
					},
				},
			},
			OsProfile: &compute.OSProfile {
				AdminUsername: to.StringPtr(adminName),
				AdminPassword: to.StringPtr(adminPassword),
				ComputerName: to.StringPtr(vmName),
				WindowsConfiguration: &compute.WindowsConfiguration { ProvisionVMAgent: to.BoolPtr(true) },
			},
		},
	}	
	
	if _,err := vmc.CreateOrUpdate(groupName, vmName, vmParams); err != nil {
		return fmt.Errorf("Failed to create virtual machine '%s' in location '%s': '%s'\n", vmName, *group.Location, err.Error())
	}
	
	return nil
}
