# Creating a Service and Virtual Machine

## Introduction

This sample uses the legacy APIs to create a hosted service and then a Linux virtual machine within it. It relies on
management certificates for authentication, which is very different from how the newer Azure Resource Manager
APIs do it. 

The client currently supports authentication to the Service Management API with certificates or Azure `.publishSettings` file.
You can download the `.publishSettings` file for your subscriptions [here](https://manage.windowsazure.com/publishsettings).

## The Code

The first step is just go gather a few pieces of information -- the VM parameters, the OS image to use, etc. Credentials
are established by loading the management certificate in the `.publishSettings` file.

```go
func main() {
    
    dnsName := "test-vm-from-go"
    storageAccount := "gosdktest"
    location := "West US"
    vmSize := "Small"
    vmImage := "b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04-LTS-amd64-server-20140724-en-us-30GB"
    userName := "testuser"
    userPassword := "Test123"

    client, err := management.ClientFromPublishSettingsFile("<<a path to a .publishsettings file>>", "")
    if err != nil {
        panic(err)
    }
```

Then, we have to have a service to put the virtual machine within. Since Azure will treat an attempt to create
a service with the same name as one that exists, we need to look for a `ConflictError` specifically and ignore
it. We may be adding a second (or third, etc.) virtual machine to an existing service, so the error is benign.



```go
    // create hosted service
    if err := hostedservice.NewClient(client).CreateHostedService(hostedservice.CreateHostedServiceParameters{
        ServiceName: dnsName,
        Location:    location,
        Label:       base64.StdEncoding.EncodeToString([]byte(dnsName))}); err != nil {
        
        if azErr,ok := err.(management.AzureError); (!ok || azErr.Code != "ConflictError") {
            panic(err)            
        }
    }
```

The third step is to create a VM configuration.

```go
    // create virtual machine
    role := vmutils.NewVMConfiguration(dnsName, vmSize)
    vmutils.ConfigureDeploymentFromPlatformImage(
        &role,
        vmImage,
        fmt.Sprintf("http://%s.blob.core.windows.net/sdktest/%s.vhd", storageAccount, dnsName),
        "")
    vmutils.ConfigureForLinux(&role, dnsName, userName, userPassword)
    vmutils.ConfigureWithPublicSSH(&role)
```

Lastly, we just create the virtual machine using the configuration details and wait for the result. Creating
a virtual machine may take a while, so the operation is an asynchronous REST API and requires clients to explicitly poll
for its completion.

```go
    operationID, err := virtualmachine.NewClient(client).
        CreateDeployment(role, dnsName, virtualmachine.CreateDeploymentOptions{})
    if err != nil {
        panic(err)
    }
    if err := client.WaitForOperation(operationID, nil); err != nil {
        panic(err)
    }
}
```