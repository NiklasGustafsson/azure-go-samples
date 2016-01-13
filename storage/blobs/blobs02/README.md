# Uploading an Image File

This sample demonstrates a fairly simple blob operation -- uploading a JPEG image. We assume that you have already gone through the blobs01
sample and won't repeat any of the comments on code that is duplicated here. Rather than coming up with a random container and blob name, we
will fix the container name and take the blob name from the command line.

Invoking the sample looks like this (after you call 'go install' to have the binary created, of course):

```
    bin/blobs02 image-file-path blob-name.jpg 
```
And the code to pick up the arguments:
```go
    if len(os.Args) < 3 {
        fmt.Printf("usage: blob02 file-name blob-name\n")
        return
    }
    
    fileName := os.Args[1]
    blob := os.Args[2]	
```
Once the container has been created, we'll look for the file, making sure it's available. Using a defer
statement makes sure that it is closed afterwards. It would be, anyway, since this is a main() function,
but it's a good practice to always include it. We also need the size of the file, so call for the file info.
```go
    f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("ERROR: Unable to locate or open file at %s (%s)\n", fileName, err.Error())
        return
	}

    defer f.Close()  

    fileInfo, err := f.Stat()
    
    if err != nil {
		fmt.Printf("ERROR: Unable to retrieve info on file at %s (%s)\n", fileName, err.Error())
        return        
    }
```
Once we have that, it's time to create the blob by passing in the file to the blob creation function.
The function takes a blob properties struct pointer (nil is valid), the only way we have of setting the
content type of the file on the server. This is important so that a client downloading it will get the
MIME type right.

After the file is uploaded, there's a redundant call to SetBlobProperties. This is an alternative way
of setting the blob's content type, used for demonstration purposes only.
```go
	props := storage.BlobProperties{ContentType: imageJPG}
	
	if err := cli.CreateBlockBlobFromReader(cnt, blob, uint64(fileInfo.Size()), f, &props); err != nil {
		fmt.Printf("Failed to create blob '%s' in  '%s': %s\n", blob, cnt, err.Error())
		return
	}
    url := cli.GetBlobURL(cnt,blob)
    
	fmt.Printf("Successfully uploaded file to '%s'\n", url)
    
	if err := cli.SetBlobProperties(cnt, blob, props); err != nil {
		fmt.Printf("Failed to set properties for '%s': %s\n", url, err.Error())
		return
	}    
```
Finally, we'll call the server to see what properties it has stored for the image. 
```go
    stored,err := cli.GetBlobProperties(cnt,blob)
    if err != nil {
		fmt.Printf("Failed to retrieve blob properties for '%s': %s\n", url, err.Error())
		return
    }    
    
    fmt.Printf("Stored properties: %s\n", helpers.ToJSON(*stored))
```
