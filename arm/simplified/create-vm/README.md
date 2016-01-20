# Creating a VM Using Simplified ARM APIs

The Go SDK offers a limited number of "simplified" APIs to make the job of many everyday tasks easier. How many copies
of the code in the ARM VM creation sample do we need in the world, after all? Most Linux or Windows VMs need a fixed
number of things -- a NIC, a disk, a public IP address, etc. In many cases, all we really need to do as programmers
is come up with pithy names for things.

In this sample, we'll take a look at the API to create a virtual machine from a set of configuration options. We'll need
to come up with a user name, a password, and a name of the storage account in which a disk is allocated for the VM.

## The Code Flow

The sample consists of two calls to 'CreateSimpleVM(),' the simplified API in question. As usual, our code needs to 
authenticate against the Azure Active Directory before we can do anything.

It's going to take a while to get the virtual machine started. Monitor its progress in the [Azure Portal](http://portal.azure.com)
and when it's done, connect to it using Remote Desktop.

As is the case with all ARM samples here, the easiest way to clean up afterwards is to simply delete the entire resource group.

After that, it's really just a matter of coming up with the right image reference data and names for things. In the first
call, we'll create a Windows Server 2012 VM.

```go
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
```

In the second, which is remarkably similar to the first, we'll create an Ubuntu 14.04 VM.

```go
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
```

That's it! Of course, these simplification comes at a cost -- the ability to create complex configurations for VMs
is not there, it's all vanilla. The best way to deal with complex configurations is to use ARM templates, though, not
to create resources individually by orchestrating it all from a client.