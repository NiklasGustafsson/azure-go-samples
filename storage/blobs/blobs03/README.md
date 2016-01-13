# Creating a page blob

This sample will create and write to a page blob, which are blobs optimized for random-access patterns. Unlike block and append blobs,
page blobs are always a multiple of 512 bytes long, and data can only be written in n*512-byte chunks, aligned on 512-byte boundaries.
Each such 512-byte chunk is called *a page.*

We assume that you have already gone through the blobs01 sample and won't repeat any of the comments on code that is duplicated here.

Invoking the sample looks like this (after you call 'go install' to have the binary created, of course):

```
    bin/blobs03 <<blob-name>> 
```
The code to pick up the arguments:
```go
	if len(os.Args) < 2 {
		fmt.Printf("usage: blob03 blob-name\n")
		return
	}
	
	blob := os.Args[1]
```
The size of the blob has to be defined when it is created, and since it must be a multiple of 512 bytes, we define
a constant for it.
```go
const blobSize = 8*512	// Each page is 512 bytes long.
```
Once the container has been created (not shown), the first action is to create an empty page blob. This is analogous to
creating a Go byte array using 'make()' -- the blob is zero-initialized and immediately ready to use.
```go
	if err := cli.PutPageBlob(cnt, blob, blobSize); err != nil {
		fmt.Printf("Failed to create '%s' in  '%s': %s\n", blob, cnt, err.Error())
		return
	}
```
With an empty blob having been created, it's time to write to it. In this example, we will write one page at a time,
but it's not necessary to make separate calls, you can write several contiguous pages at once.
```go
	if err := writePage(cli, blob, 0, 511, data[0:512]); err != nil {
		return
	}
	if err := writePage(cli, blob, 512, 1023, data[512:1024]); err != nil {
		return	
	}
	if err := writePage(cli, blob, 1024, 1535, data[1024:1536]); err != nil {
		return	
	}
```
The page blob update logic is kept in a function: 
```go
func writePage(cli storage.BlobStorageClient, name string, startByte, endByte int64, chunk []byte) error {
	
	if err := cli.PutPage(cnt, name, startByte, endByte, storage.PageWriteTypeUpdate, chunk); err != nil {
		url := cli.GetBlobURL(cnt,name)
		fmt.Printf("Failed to write page to %s: %s\n", url, err.Error())
		return err
	}
	return nil	
}
```
At this point, you can use an external tool such as [Postman](https://www.getpostman.com/) to look at the data
that's been uploaded. To add proper programmatic validation, we add a function to retrieve the data and check
it against what we're expecting.
It calls 'GetBlob()' to access the data stream, reads what is sent back, and compares it against what we're
expecting to see. 
```go
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
```
The calls to the validation function look like this. The first three pages should correspond to what was written,
the next five should be all zero.
```go
	if err := validate(cli, blob, 0, 1535, data[0:1536]); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}

	if err := validate(cli, blob, 1536, blobSize-1, make([]byte,blobSize)); err != nil {
		fmt.Printf("Failed validation of %s: %s\n", url, err.Error())
		return
	}
```
Azure page blobs allow you to clear, that is, set to all zeroes, a range of pages. The sample shows how to do this in the 
clearPage() function. It comes down to passing in a different kind of page write type value, and no data block, as shown below.

```go
	err := cli.PutPage(cnt, name, startByte, endByte, storage.PageWriteTypeClear, nil)
```

Once we have a page blob in Azure, we can check which pages contain data and which do not (zeroes are data, too, but
you know what we mean!) by calling the 'GetPageRanges()' function. It returns a list of ranges of pages that have been written
to rather than cleared or left in their initial state:

```go
	ranges,err := cli.GetPageRanges(cnt,blob)
	...	
	for _,rng := range ranges.PageList {
		fmt.Printf("[%d,%d]\n", rng.Start, rng.End)
	}
```

That's pretty much it for operations specific to page blobs -- create, write pages, clear pages, examine ranges. 