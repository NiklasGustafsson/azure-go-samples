# Basic Blob Storage Uses

This simplistic sample demonstrates how to get started with Azure Blob Storage and the Go SDK. It will use storage credentials to 
create a client, use the client to first create a container and an empty block blob. 

First, we need two pieces of information to connect to a storage account -- it's name, and one of the account keys (primary or secondary).
You can find these on the Azure portal and copy them here. Doing so is clearly not a secure approach, but proper use of various platform-specific
private key infrastructure (PKI) mechanisms to store credentials is beyond the scope of this sample.
 
```go
const accountName = "<<INSERT ACCOUNT NAME HERE>>"
const accountKey  = "<<INSERT ACCOUNT KEY HERE>>"
```
Once we have the credentials, we'll make up a container name and create it if it does not already exist. Since creating a container that
already exists may or may not be benign, depending on your application, the SDK offers both idempotent and non-idempotent container
creation APIs. This is the idempotent one.

```go
	cnt := containerPrefix + helpers.RandString(32-len(containerPrefix))
	
	cli := helpers.GetStorageClient(accountName, accountKey).GetBlobService()

	ok, err := cli.CreateContainerIfNotExists(cnt, storage.ContainerAccessTypePrivate)   
	if !ok {
		fmt.Printf("Failed to create container '%s': %s\n", cnt, err.Error())
		return
	}
```
Since this is a test, we should make sure that the container is removed afterwards. In most real applications, you wouldn't do this, and if
you want to take a look at the data after running the sample, this should be commented out.
```go
	defer func() { 
			fmt.Printf("Deleting '%s' and all its contents\n", cnt)
			cli.DeleteContainer(cnt) 
		}() 
```
Make up a name and use it to create an empty block blob. 
```go
	blob := blobPrefix + helpers.RandString(32-len(blobPrefix))
	
	if err := cli.CreateBlockBlob(cnt, blob); err != nil {
		fmt.Printf("Failed to create blob '%s' in  '%s': %s\n", blob, cnt, err.Error())
		return
	}
```

