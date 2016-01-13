package main

import (
	"os"
	"fmt"
	"io/ioutil"
		
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/azure-go-samples/helpers"
)

const cnt = "blob03-container"

const accountName = "<<INSERT ACCOUNT NAME HERE>>"
const accountKey  = "<<INSERT ACCOUNT KEY HERE>>"

const blobSize = 8*512	// Each page is 512 bytes long.

func main() {
	
	if len(os.Args) < 2 {
		fmt.Printf("usage: blob03 blob-name\n")
		return
	}
	
	blob := os.Args[1]
	
	cli := helpers.GetStorageClient(accountName, accountKey).GetBlobService()
		
	// Create a container.
	
	_, err := cli.CreateContainerIfNotExists(cnt, storage.ContainerAccessTypeBlob)
	if err != nil {
		fmt.Printf("ERROR: Failed to create container '%s': %s\n", cnt, err.Error())
		return
	}

	fmt.Printf("Successfully created '%s'\n", cnt)
	
	// First, create an empty page blob
	if err := cli.PutPageBlob(cnt, blob, blobSize); err != nil {
		fmt.Printf("Failed to create '%s' in  '%s': %s\n", blob, cnt, err.Error())
		return
	}
	
	// Then, write some pages. We will write pages 0, 1, and 2, leaving five pages blank.
	
	data := helpers.RandBytes(1536)
	
	if err := writePage(cli, blob, 0, 511, data[0:512]); err != nil {
		return
	}
	if err := writePage(cli, blob, 512, 1023, data[512:1024]); err != nil {
		return	
	}
	if err := writePage(cli, blob, 1024, 1535, data[1024:1536]); err != nil {
		return	
	}
	
	// Now that we have written the pages, let's check what's been stored.
	
	url := cli.GetBlobURL(cnt,blob)

	// The first three pages should correspond to what was written.
	if err := validate(cli, blob, 0, 1535, data[0:1536]); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}

	// The next five pages should be all zeroes.
	if err := validate(cli, blob, 1536, blobSize-1, make([]byte,blobSize)); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}
	
	fmt.Printf("Successfully created, wrote to, and read a page blob at '%s\n", url)

	// Next, let's try clearing a page, that is, setting it to all zeroes.
	
	if err := clearPage(cli, blob, 512, 1023); err != nil {
		return	
	}
	
	// Now that we have written the pages, let's check what's been stored.
	
	// The first and third pages should correspond to what was written.
	if err := validate(cli, blob, 0, 511, data[0:512]); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}
	// The second page should be all zeroes.
	if err := validate(cli, blob, 512, 1023, make([]byte,1024)); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}
	if err := validate(cli, blob, 1024, 1535, data[0:1536]); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}
	
	// The last five pages should still be all zeroes.
	if err := validate(cli, blob, 1536, blobSize-1, make([]byte,blobSize)); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}
	
	// We can check which pages have data by getting the valid page ranges.
	
	ranges,err := cli.GetPageRanges(cnt,blob)
	if err != nil {
		fmt.Printf("Failed to get page ranges for %s: %s\n", url, err.Error())
		return	
	}
	
	fmt.Printf("Successfully cleared a page at '%s\n", url)

	for _,rng := range ranges.PageList {
		fmt.Printf("[%d,%d]\n", rng.Start, rng.End)
	}
	
	pl := ranges.PageList
	
	if len(pl) != 2 || pl[0].Start != 0 || pl[0].End != 511 || pl[1].Start != 1024 || pl[1].End != 1535 {
		fmt.Printf("Incorrect page list returned\n")
		return
	}

	fmt.Printf("Successfully checked page ranges of '%s\n", url)
}

func writePage(cli storage.BlobStorageClient, name string, startByte, endByte int64, chunk []byte) error {
	
	if err := cli.PutPage(cnt, name, startByte, endByte, storage.PageWriteTypeUpdate, chunk); err != nil {
		url := cli.GetBlobURL(cnt,name)
		fmt.Printf("Failed to write pages to %s: %s\n", url, err.Error())
		return err
	}
	return nil	
}

func clearPage(cli storage.BlobStorageClient, name string, startByte, endByte int64) error {
	
	if err := cli.PutPage(cnt, name, startByte, endByte, storage.PageWriteTypeClear, nil); err != nil {
		url := cli.GetBlobURL(cnt,name)
		fmt.Printf("Failed to clear pages of %s: %s\n", url, err.Error())
		return err
	}
	return nil	
}

func validate(cli storage.BlobStorageClient, blob string, startByte, endByte int64, data []byte) error {
	
	url := cli.GetBlobURL(cnt,blob)

	reader,err := cli.GetBlob(cnt, blob)
	if err != nil {
		return fmt.Errorf("Failed to read from %s: %s\n", url, err.Error())
	}
	
	defer reader.Close()

	dataRead,err := ioutil.ReadAll(reader)
	
	if err != nil {
		return fmt.Errorf("Failed to read from %s: %s\n", url, err.Error())
	}	
	
	same := true
	for i := startByte; i <= endByte; i++ {
		if data[i] != dataRead[i] {
			same = false
		}		
	}
	
	if !same {
		return fmt.Errorf("Failed to read data properly from %s: %s\n", url, err.Error())
	}
	
	return nil	
}