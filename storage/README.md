# Azure Storage

The following samples are meant to highlight how Azure storage may be used for a variety of application
storage scenarios.

[Creating a Block Blob](./blobs01)

This sample will create a blob container, and then place an empty, private, block blob within it.

[Uploading an Image](./blobs02)

This sample creates a container for images and uploads a file to  it, setting its MIME type and checking
that the stored data is correct.

[Creating a Page Blob](./blobs03)

Here, we create a page blob, which is convenient for certain scenarios that require random-access control
over contents within the blob. Parts of the blob are filled with random data, others with zeroes. 