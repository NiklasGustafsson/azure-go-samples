// See README.md file for extensive commentary on the sample code in this file.

package main

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/azure-go-samples/helpers"
)
const containerPrefix = "cont-"
const blobPrefix = "blob-"

const accountName = "<<INSERT ACCOUNT NAME HERE>>"
const accountKey  = "<<INSERT ACCOUNT KEY HERE>>"

func main() {
    
	cnt := containerPrefix + helpers.RandString(32-len(containerPrefix))
	
	cli := helpers.GetStorageClient(accountName, accountKey).GetBlobService()
	
	// Create a container.
	
	ok, err := cli.CreateContainerIfNotExists(cnt, storage.ContainerAccessTypePrivate)   
	if !ok {
		fmt.Printf("Failed to create container '%s': %s\n", cnt, err.Error())
		return
	}

	// Make sure not to clobber up the storage account.
	
	defer func() { 
			fmt.Printf("Deleting '%s' and all its contents\n", cnt)
			cli.DeleteContainer(cnt) 
		}() 

	fmt.Printf("Successfully created '%s'\n", cnt)
	
	// Create an empty blob
	
	blob := blobPrefix + helpers.RandString(32-len(blobPrefix))
	
	if err := cli.CreateBlockBlob(cnt, blob); err != nil {
		fmt.Printf("Failed to create blob '%s' in  '%s': %s\n", blob, cnt, err.Error())
		return
	}

	fmt.Printf("Successfully created '%s'\n", blob)
}
