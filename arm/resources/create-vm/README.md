# Creating a Windows Server VM Using Azure Resource Manager APIs

This sample builds on the 'check-name' and 'create-group' ARM samples, which demonstrated how to authenticate
and create a resource group, respectively. Now, it's time to kick it up a notch and do something real, create
a virtual machine within a resource group that we will create.

One thing that will stand out is how much more complex the ARM APIs are than the classic API to create a virtual
machine. This is due to the significantly increased flexibility that ARM provides, and we're examining the low-
level mechanisms here. There is a more efficient way to do the same thing with ARM templates, but it is worth
first understanding all the gory details of creating resources and linking them together.

## The Code Flow

Let's dive right into the code. The main() function doesn't really do any of the real work, it's all been
broken out into its stages, but the logic provides a nice outline of what needs to be done:

1. Create a resource group
2. Create a storage account
3. Create an availability set
4. Create a virtual network and a subnet
5. Create a network interface (NIC) with a public IP address, if applicable.
6. Create the virtual machine

The resources in 1-4 can be shared with other virtual machines, while the resources created in stage 5 are
per-virtual machine the network interface and its associated public IP address are specific to the virtual machine.
To create more than one virtual machine, you have to create another network interface (and public address, if there
should be one).

**Note**: all error-checking code has been removed from the code snippets here, but it is present in the
actual sample source code.

## The Functions
**createResourceGroup()**

This is more or less a repeat of the 'create-group' sample, except that we also register ARM resource providers
that we will be needing later.
```go
	rgc := resources.NewGroupsClient(subscription)
	rgc.Authorizer = spt	
	
	params := resources.ResourceGroup { ame:&name,Location:&location }
	
	group,err = rgc.CreateOrUpdate(name, params)
	
	fmt.Printf("Created resource group '%s'\n", *group.Name)
	rpc := resources.NewProvidersClient(subscription)
	rpc.Authorizer = spt
	
	_,err1 := rpc.Register("Microsoft.Storage")
	...
```
**createStorageAccount()**

This function is similar to the 'check-name' sample, except that it not only checks to see whether a name
is available, it also creates the account if it is. This function takes a while to finish, since creation of a 
storage account is an expensive process and can take a minute or so.
```go
	cna, err := ac.CheckNameAvailability(...)

	...
	
	if to.Bool(cna.NameAvailable) {
		
		name := *group.Name
		props := storage.AccountPropertiesCreateParameters {
			AccountType: storage.StandardLRS,
		}
		
		_,err = ac.Create(name, name,
			storage.AccountCreateParameters {
				Name: group.Name,
				Location: group.Location,
				Properties: &props,
			})
	}
```
**createAvailabilitySet()**

Strictly speaking, it doesn't make much sense to use an [availability set](https://azure.microsoft.com/en-us/documentation/articles/virtual-machines-manage-availability/)
for a single VM, but most real installations will use more than one virtual machine, and so the code demonstrates how to set it up.

The availability set function is the simplest in the sample. The name is created from the group name, which
simplifies the code a bit. In a real installation, there would be multiple availability sets with fixed names,
typically one for each tier in your service architecture. The name and location of the set are the only parameters
needed to create the set.
```go
	avsc := compute.NewAvailabilitySetsClient(subscription)
	avsc.Authorizer = spt; 
	
	name := *group.Name

	result,err = avsc.CreateOrUpdate(
		name,
		name+"avset",
		compute.AvailabilitySet {
			Location: group.Location,
		})
```  

**createNetwork()**

For virutal machines to find each other within the data center, they need to be configured on the same virtual network.
We need to create a vnet and a subnet in order to get everything set up correctly. The details of network configuration
are beyond the scope of this tutorial sample, we're concerned with setting up the right calls only.

The first step is to create the clients, in what should not be a familiar way.
```go
	vnetc := network.NewVirtualNetworksClient(subscription)
	snetc := network.NewSubnetsClient(subscription)
	
	vnetc.Authorizer = spt; 
	snetc.Authorizer = spt; 
```
Then, we set up some of the properties data structures that will contain the arguments to the service calls.
```go
	snet := network.Subnet {
		Name: &subnet, 
		Properties: &network.SubnetPropertiesFormat {
			AddressPrefix: to.StringPtr("10.0.0.0/24"),
		},
	}
	snets := make([]network.Subnet, 1, 1)
	snets[0] = snet
	
	addrPrefixes := make([]string,1,1)
	addrPrefixes[0] = "10.0.0.0/16"
	address := network.AddressSpace {
		AddressPrefixes: &addrPrefixes,
	}
```
Finally, our code makes the two service calls that we need, one for the vnet, one for the subnet. The subnet is the only data
we'll need to create the virtual machine, so that's what the function returns.
```go	
	nwkProps := network.VirtualNetworkPropertiesFormat {
		AddressSpace: &address,
		Subnets: &snets,
	}
	
	_,err = vnetc.CreateOrUpdate(
		name,
		vnet,
		network.VirtualNetwork {
			Location: group.Location,
			Properties: &nwkProps,
		})
	
	snetResult, err = snetc.CreateOrUpdate(name, vnet, subnet, snet)
```
**createNetworkInterface()**

With a virtual network and a subnet available, we can now proceed to set up the network interfaces that our virtual machine will
be using to communicate. In our case, we want to be able to communicate with the Windows server from other locations, so we also 
need to give the interface a public IP address. The public IP address is not necessary for virtual machines that only need to be reachable from
other machines on the same virtual network.

Once more, we'll need more than one client to get anything done:
```go
	pipc  := network.NewPublicIPAddressesClient(subscription)
	nicc  := network.NewInterfacesClient(subscription)

	pipc.Authorizer = spt; 
	nicc.Authorizer = spt; 
```

With those in hand, we can set up the public IP address, a fairly simple call:

```go
	pipResult,err := pipc.CreateOrUpdate(
		groupName, 
		ipName, 
		network.PublicIPAddress {
			Location: group.Location, 
			Properties: &network.PublicIPAddressPropertiesFormat {
				PublicIPAllocationMethod: network.Dynamic,
			},
		})
``` 
and then the network interface itself, which requires a little more parameter setup. The network interface is what we'll need
in order to create a virtual machine, so that's what the function returns.
```go
	nicProps := network.InterfaceIPConfigurationPropertiesFormat { 
		PublicIPAddress: &network.SubResource { ID: pipResult.ID }, 
		Subnet: &network.SubResource { ID: subnet.ID },
	}
		
	ipConfigs := make([]network.InterfaceIPConfiguration,1,1)
	ipConfigs[0] = network.InterfaceIPConfiguration {
		Name: to.StringPtr(nicName+"Config"),
		Properties: &nicProps,
	}
	props := network.InterfacePropertiesFormat { IPConfigurations: &ipConfigs }
		
	networkInterface,err = nicc.CreateOrUpdate(
		groupName, 
		nicName, 
		network.Interface{
			Location: group.Location, 
			Properties: &props,
		})	
```
**createVirtualMachine()**

The actual virtual machine creation call is fairly straight-forward, but there are a lot of parameters that need to be set up first.
VM creation takes a VirtualMachine object containing a number of distinct 'profiles,' describing what OS image to use, how big a VM to
create, what virtual hard drives to mount, and where to allocate backing storage (this is what the storage account is needed for), 
as well as how to configure the machine with an admin user and password. The network interface and availability set are also
identified.

Since a virtual machine may have more than one network interface, the NIC data needs to be passed in a slice:

```go
	netRefs := make([]compute.NetworkInterfaceReference,1,1)
	netRefs[0] = compute.NetworkInterfaceReference{ ID: networkInterface.ID } 
```
The rest is just a big document:
```go
	vmParams := compute.VirtualMachine { 
		Location: group.Location,
		Properties: &compute.VirtualMachineProperties { 
			AvailabilitySet: &compute.SubResource { ID: availSet.ID },
			HardwareProfile: &compute.HardwareProfile { VMSize: compute.StandardA0 },
			NetworkProfile: &compute.NetworkProfile { NetworkInterfaces: &netRefs },
			StorageProfile: &compute.StorageProfile { 
				ImageReference: &compute.ImageReference {
					Publisher: to.StringPtr("MicrosoftWindowsServer"),
					Offer: to.StringPtr("WindowsServer"),
					Sku: to.StringPtr("2012-R2-Datacenter"),
					Version: to.StringPtr("latest"),
				}, 
				OsDisk: &compute.OSDisk {
					Name: to.StringPtr("mytestod1"),
					CreateOption: compute.FromImage,
					Vhd: &compute.VirtualHardDisk {
						URI: to.StringPtr("http://" + accountName + ".blob.core.windows.net/vhds/mytestod1.vhd") },
				},
			},
			OsProfile: &compute.OSProfile {
				AdminUsername: to.StringPtr(adminName),
				AdminPassword: to.StringPtr(adminPassword),
				ComputerName: to.StringPtr(vmName),
				WindowsConfiguration: &compute.WindowsConfiguration {
					ProvisionVMAgent: to.BoolPtr(true),
				},
			},
		},
	}	
```

Compared to all that, the actual call is trivial:

```go
	_,err := vmc.CreateOrUpdate(groupName, vmName, vmParams)
```
It's going to take a while to get the virtual machine started. Monitor its progress in the [Azure Portal](http://portal.azure.com)
and when it's done, connect to it using Remote Desktop.

As is the case with all ARM samples here, the easiest way to clean up afterwards is to simply delete the entire resource group.