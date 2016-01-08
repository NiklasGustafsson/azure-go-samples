package main

import (
    "encoding/base64"
    "fmt"

    "github.com/Azure/azure-sdk-for-go/management"
    "github.com/Azure/azure-sdk-for-go/management/hostedservice"
    "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
    "github.com/Azure/azure-sdk-for-go/management/vmutils"
)

func main() {
    dnsName := "test-vm-from-go"
    storageAccount := "gosdktest"
    location := "West US"
    vmSize := "Small"
    vmImage := "b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04-LTS-amd64-server-20140724-en-us-30GB"
    userName := "testuser"
    userPassword := "Test123"

    client, err := management.ClientFromPublishSettingsFile("/Users/niklasg/Downloads/Internal-credentials.publishsettings", "")
    if err != nil {
        panic(err)
    }

    // create hosted service
    if err := hostedservice.NewClient(client).CreateHostedService(hostedservice.CreateHostedServiceParameters{
        ServiceName: dnsName,
        Location:    location,
        Label:       base64.StdEncoding.EncodeToString([]byte(dnsName))}); err != nil {
        
        if azErr,ok := err.(management.AzureError); (!ok || azErr.Code != "ConflictError") {
            panic(err)            
        }
    }

    // create virtual machine
    role := vmutils.NewVMConfiguration(dnsName, vmSize)
    vmutils.ConfigureDeploymentFromPlatformImage(
        &role,
        vmImage,
        fmt.Sprintf("http://%s.blob.core.windows.net/sdktest/%s.vhd", storageAccount, dnsName),
        "")
    vmutils.ConfigureForLinux(&role, dnsName, userName, userPassword)
    vmutils.ConfigureWithPublicSSH(&role)

    operationID, err := virtualmachine.NewClient(client).
        CreateDeployment(role, dnsName, virtualmachine.CreateDeploymentOptions{})
    if err != nil {
        panic(err)
    }
    if err := client.WaitForOperation(operationID, nil); err != nil {
        panic(err)
    }
}