// See README.md file for extensive commentary on the sample code in this file.

package main

import (
	"os"
	"fmt"
	
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/azure-go-samples/helpers"
)
const cnt = "blob02-container"
const imageJPG = "image/jpeg"

const accountName = "<<INSERT ACCOUNT NAME HERE>>"
const accountKey  = "<<INSERT ACCOUNT KEY HERE>>"

func main() {
	
	if len(os.Args) < 3 {
		fmt.Printf("usage: blob02 file-name blob-name\n")
		return
	}
	
	fileName := os.Args[1]
	blob := os.Args[2]	

	cli := helpers.GetStorageClient(accountName, accountKey).GetBlobService()
	
	// Create a container.
	
	_, err := cli.CreateContainerIfNotExists(cnt, storage.ContainerAccessTypeBlob)   
	if err != nil {
		fmt.Printf("ERROR: Failed to create container '%s': %s\n", cnt, err.Error())
		return
	}

	fmt.Printf("Successfully created '%s'\n", cnt)
	
	// Open the file
	
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("ERROR: Unable to locate or open file at %s (%s)\n", fileName, err.Error())
		return
	}

	defer f.Close()  

	// Get the file info, so we can know its size
	
	fileInfo, err := f.Stat()
	
	if err != nil {
		fmt.Printf("ERROR: Unable to retrieve info on file at %s (%s)\n", fileName, err.Error())
		return
	}
	
	props := storage.BlobProperties{ContentType: imageJPG}
	
	// Create the blob from the file. Also, pass in a properties block so that the
	// content type may be set.
	if err := cli.CreateBlockBlobFromReader(cnt, blob, uint64(fileInfo.Size()), f, &props); err != nil {
		fmt.Printf("Failed to create '%s' in  '%s': %s\n", blob, cnt, err.Error())
		return
	}
	
	// We'll want the url of the blob
	url := cli.GetBlobURL(cnt,blob)
	
	fmt.Printf("Successfully uploaded file to '%s'\n", url)
	
	if err := cli.SetBlobProperties(cnt, blob, props); err != nil {
		fmt.Printf("Failed to set properties for '%s': %s\n", url, err.Error())
		return
	}
	
   	fmt.Printf("Successfully set properties for '%s'\n", url)

	// Just to make sure, let's see what the properties on the server!
	stored,err := cli.GetBlobProperties(cnt,blob)
	if err != nil {
		fmt.Printf("Failed to retrieve blob properties for '%s': %s\n", url, err.Error())
		return
	}	
	
	fmt.Printf("Stored properties: %s\n", helpers.ToJSON(*stored))
}
