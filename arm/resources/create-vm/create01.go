package main

import (
	"fmt"

	"github.com/Azure/azure-go-samples/helpers"
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/azure-sdk-for-go/arm"
	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
)

func main() {

	groupName := "createvm01"
	groupLocation := "West US"

	var group resources.ResourceGroup
	var subnet network.Subnet
	var nic network.Interface
	var avset compute.AvailabilitySet

	client, err := helpers.AuthenticateForARM()
	if err != nil {
		fmt.Printf("Failed to authenticate: '%s'\n", err.Error())
		return
	}

	client.RequestInspector = helpers.WithInspection()
	client.ResponseInspector = helpers.ByInspecting()

	group, err = createResourceGroup(groupName, groupLocation, client)

	if err != nil {
		fmt.Printf("ERROR:'%s'\n", err.Error())
		return
	}

	if err := createStorageAccount(group, client); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return
	}

	if avset, err = createAvailabilitySet(group, client); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return
	}

	if subnet, err = createNetwork(group, client); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return
	}

	if nic, err = createNetworkInterface("01", group, subnet, client); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return
	}

	if err := createVirtualMachine(group, "vm001", "admin", "foobar1234", avset, nic, client); err != nil {
		fmt.Printf("ERROR: '%s'\n", err.Error())
		return
	}
}

func createResourceGroup(
	name, location string,
	arm arm.Client) (group resources.ResourceGroup, err error) {

	rgc := arm.ResourceGroups()
	rpc := arm.Providers()

	params := resources.ResourceGroup{Name: &name, Location: &location}

	group, err = rgc.CreateOrUpdate(name, params)
	if err != nil {
		err = fmt.Errorf("Failed to create resource group '%s' in location '%s': '%s'\n", name, location, err.Error())
		return
	}

	fmt.Printf("Created resource group '%s'\n", *group.Name)

	if _, err1 := rpc.Register("Microsoft.Storage"); err != nil {
		err = fmt.Errorf("Failed to register resource provider 'Microsoft.Storage': '%s'\n", err1.Error())
	}
	if _, err1 := rpc.Register("Microsoft.Network"); err != nil {
		err = fmt.Errorf("Failed to register resource provider 'Microsoft.Network': '%s'\n", err1.Error())
	}
	if _, err1 := rpc.Register("Microsoft.Compute"); err != nil {
		err = fmt.Errorf("Failed to register resource provider 'Microsoft.Compute': '%s'\n", err1.Error())
	}

	return
}

func createStorageAccount(
	group resources.ResourceGroup,
	arm arm.Client) error {

	ac := arm.StorageAccounts()

	cna, err := ac.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: group.Name,
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})

	if err != nil {
		return err
	}

	if to.Bool(cna.NameAvailable) {

		name := *group.Name
		props := storage.AccountPropertiesCreateParameters{AccountType: storage.StandardLRS}

		_, err = ac.Create(name, name,
			storage.AccountCreateParameters{
				Location:   group.Location,
				Properties: &props,
			})

		if err != nil {
			return fmt.Errorf("Failed to create storage account '%s' in location '%s': '%s'\n", name, *group.Location, err.Error())
		}
	}

	return nil
}

func createAvailabilitySet(
	group resources.ResourceGroup,
	arm arm.Client) (result compute.AvailabilitySet, err error) {

	avsc := arm.AvailabilitySets()

	name := *group.Name

	result, err = avsc.CreateOrUpdate(name, name+"avset", compute.AvailabilitySet{Location: group.Location})
	if err != nil {
		err = fmt.Errorf("Failed to create availability set '%s' in location '%s': '%s'\n", name, *group.Location, err.Error())
		return
	}

	return result, nil
}

func createNetwork(
	group resources.ResourceGroup,
	arm arm.Client) (snetResult network.Subnet, err error) {

	vnetc := arm.VirtualNetworks()
	snetc := arm.Subnets()

	name := *group.Name
	vnet := name + "vnet"
	subnet := name + "subnet"

	snet := network.Subnet{
		Name:       &subnet,
		Properties: &network.SubnetPropertiesFormat{AddressPrefix: to.StringPtr("10.0.0.0/24")}}
	snets := make([]network.Subnet, 1, 1)
	snets[0] = snet

	addrPrefixes := make([]string, 1, 1)
	addrPrefixes[0] = "10.0.0.0/16"
	address := network.AddressSpace{AddressPrefixes: &addrPrefixes}

	nwkProps := network.VirtualNetworkPropertiesFormat{AddressSpace: &address, Subnets: &snets}

	_, err = vnetc.CreateOrUpdate(name, vnet, network.VirtualNetwork{Location: group.Location, Properties: &nwkProps})
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
	suffix string,
	group resources.ResourceGroup,
	subnet network.Subnet,
	arm arm.Client) (networkInterface network.Interface, err error) {

	pipc := arm.PublicIPAddresses()
	nicc := arm.NetworkInterfaces()

	groupName := *group.Name
	ipName := "ip" + suffix
	nicName := "nic" + suffix

	pipResult, err := pipc.CreateOrUpdate(
		groupName,
		ipName,
		network.PublicIPAddress{
			Location: group.Location,
			Properties: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.Dynamic,
			},
		})

	if err != nil {
		err = fmt.Errorf("Failed to create public ip address '%s' in location '%s': '%s'\n", ipName, *group.Location, err.Error())
		return
	}

	nicProps := network.InterfaceIPConfigurationPropertiesFormat{
		PublicIPAddress: &pipResult,
		Subnet:          &subnet}

	ipConfigs := make([]network.InterfaceIPConfiguration, 1, 1)
	ipConfigs[0] = network.InterfaceIPConfiguration{
		Name:       to.StringPtr(nicName + "Config"),
		Properties: &nicProps,
	}
	props := network.InterfacePropertiesFormat{IPConfigurations: &ipConfigs}

	networkInterface, err = nicc.CreateOrUpdate(
		groupName,
		nicName,
		network.Interface{
			Location:   group.Location,
			Properties: &props,
		})
	if err != nil {
		err = fmt.Errorf("Failed to create network interface '%s' in location '%s': '%s'\n", nicName, *group.Location, err.Error())
	}

	return
}

func createVirtualMachine(
	group resources.ResourceGroup,
	vmName, adminName, adminPassword string,
	availSet compute.AvailabilitySet,
	networkInterface network.Interface,
	arm arm.Client) error {

	vmc := arm.VirtualMachines()

	netRefs := make([]compute.NetworkInterfaceReference, 1, 1)
	netRefs[0] = compute.NetworkInterfaceReference{ID: networkInterface.ID}

	groupName := *group.Name
	accountName := groupName

	vmParams := compute.VirtualMachine{
		Location: group.Location,
		Properties: &compute.VirtualMachineProperties{
			AvailabilitySet: &compute.SubResource{ID: availSet.ID},
			HardwareProfile: &compute.HardwareProfile{VMSize: compute.StandardA0},
			NetworkProfile:  &compute.NetworkProfile{NetworkInterfaces: &netRefs},
			StorageProfile: &compute.StorageProfile{
				ImageReference: &compute.ImageReference{
					Publisher: to.StringPtr("MicrosoftWindowsServer"),
					Offer:     to.StringPtr("WindowsServer"),
					Sku:       to.StringPtr("2012-R2-Datacenter"),
					Version:   to.StringPtr("latest"),
				},
				OsDisk: &compute.OSDisk{
					Name:         to.StringPtr("mytestod1"),
					CreateOption: compute.FromImage,
					Vhd: &compute.VirtualHardDisk{
						URI: to.StringPtr("http://" + accountName + ".blob.core.windows.net/vhds/mytestod1.vhd"),
					},
				},
			},
			OsProfile: &compute.OSProfile{
				AdminUsername:        to.StringPtr(adminName),
				AdminPassword:        to.StringPtr(adminPassword),
				ComputerName:         to.StringPtr(vmName),
				WindowsConfiguration: &compute.WindowsConfiguration{ProvisionVMAgent: to.BoolPtr(true)},
			},
		},
	}

	if _, err := vmc.CreateOrUpdate(groupName, vmName, vmParams); err != nil {
		return fmt.Errorf("Failed to create virtual machine '%s' in location '%s': '%s'\n", vmName, *group.Location, err.Error())
	}

	return nil
}
