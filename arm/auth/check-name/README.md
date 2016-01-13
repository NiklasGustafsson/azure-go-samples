# Authenticating with Service Principal Credentials

## Introduction

Before using any of the Azure Resource Manager (ARM) packages, you need to understand how it authenticates and
authorizes requests.

In this sample, we demonstrate how to use service principal credentials to authorize a resource management request that
requires privileges. This is applicable to so-called *unattended* authentication scenarios, that is, those situations
where there is no interaction with a user that can enter credentials. Such situations typcially arise for automation
or service-to-service calls.  

The Azure Storage APIs themselves, which are used to add and access data in blobs, tables, files, and queues,
are not used in this example. We are concerned with the storage management APIs, specifically the one that
let's use check whether an account name is available.

The Azure Resource Manager does *not* use management certificates. Instead, it relies on [OAuth2](http://oauth.net).

The regular OAuth2 flow, which is used for interactive scenarios, is described in
[Authorization Code Grant Flow](https://msdn.microsoft.com/en-us/library/azure/dn645543.aspx). You have without a doubt 
been exposed to this many times in the past, including logging into the Azure portal.

The OAuth2 flow applicable to unattended scenarios is described in [Client Credentials Grant Flow Diagram](https://msdn.microsoft.com/en-us/library/azure/dn645543.aspx). This is the flow that the Go SDK implements and which is used in this example.

Using the client credentials grant flow with Azure requires understanding and creating one or more *Service Principals*,
which are essentially application identities registered with an Azure ActiveDirectory used in your subscription. 

There are several good blog posts, such as
[Automating Azure on your CI server using a Service Principal](http://blog.davidebbo.com/2014/12/azure-service-principal.html)
and
[Microsoft Azure REST API + OAuth 2.0](https://ahmetalpbalkan.com/blog/azure-rest-api-with-oauth2/),
that describe what this means. For details on creating and authorizing Service Principals, see the MSDN articles
[Azure API Management REST API Authentication](https://msdn.microsoft.com/en-us/library/azure/5b13010a-d202-4af5-aabf-7ebc26800b3d)
and
[Create a new Azure Service Principal using the Azure portal](https://azure.microsoft.com/en-us/documentation/articles/resource-group-create-service-principal-portal/).

Creating and configuring a service principal is the most complicated and tedious part of getting started with the Go SDK.
Once you've gone through it, save the client id, client secret, and tenant id in a safe place, such as a password manager with strong encryption.
When you create a secret in the portal, it will only be shown to you once, so make sure to copy it right away. 

## The Actual Code

Now that we have the preliminaries taken care of, let's walk through the sample code.

The first step is to load the service principal credentials from somewhere. The helper function just gets it from
a plain file on disk, clearly not a secure method, but proper use of various platform-specific
private key infrastructure (PKI) mechanisms to store credentials is beyond the scope of this sample. The storage account 
name in the sample is invalid, so it will always be rejected. Come up with some unique name that consists only of numbers
and lower-case letters (no hyphens, in other words). You can use the `helpers.RandString(n)` function, for example. 

```go
func main() {
    
    name := "storage-account-name"
       
	c, err := helpers.LoadCredentials()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
    
    sid := c["subscriptionID"]
    tid := c["tenantID"]
    cid := c["clientID"]
    secret := c["clientSecret"]
```

Once we have the subscription, tenant, and client information, we can proceed to creating the account client object and
a authentication token. The latter is then attached to the client.
    
```go
	ac := storage.NewAccountsClient(sid)

	spt, err := azure.NewServicePrincipalToken(cid, secret, tid, azure.AzureResourceManagerScope)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	ac.Authorizer = spt

```

For debugging purposes, we attach inspectors, which are simply callback functions on the client. This will allow us to "spy"
on the requests that are made on our behalf, as well as their responses. The 'helpers' package is part of the SDK samples, not
the SDK itself. 

```go
	ac.RequestInspector = helpers.WithInspection()
	ac.ResponseInspector = helpers.ByInspecting()
```

All that's left is to use the account client and its credentials to make a call to the ARM service. The SDK will automatically
use the credentials to get an OAuth2 access token on our behalf. Errors are always possible, so we need to handle them correctly.
Note that the API does not treat a negative response (the name is unavailable) as an error.

```go
	cna, err := ac.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters {
			Name: to.StringPtr(name),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})

	if err != nil {
		log.Fatalf("Error: %v", err)
	} else {
		if to.Bool(cna.NameAvailable) {
			fmt.Printf("The name '%s' is available\n", name)
		} else {
			fmt.Printf("The name '%s' is unavailable because %s\n", name, to.String(cna.Message))
		}
	}
}
```

## A Few Additional Comments
 
Each ARM client composes with [autorest.Client](https://godoc.org/github.com/Azure/go-autorest/autorest#Client).
[autorest.Client](https://godoc.org/github.com/Azure/go-autorest/autorest#Client)
enables altering the behavior of the API calls by leveraging the decorator pattern of
[go-autorest](https://github.com/Azure/go-autorest). For example, in the code above, the
[azure.ServicePrincipalToken](https://godoc.org/github.com/Azure/go-autorest/autorest/azure#ServicePrincipalToken)
includes a
[WithAuthorization](https://godoc.org/github.com/Azure/go-autorest/autorest#Client.WithAuthorization)
[autorest.PrepareDecorator](https://godoc.org/github.com/Azure/go-autorest/autorest#PrepareDecorator)
that applies the OAuth2 authorization token to the request. It will, as needed, refresh the token
using the supplied credentials.

All Azure ARM API calls return an instance of the
[autorest.Error](https://godoc.org/github.com/Azure/go-autorest/autorest#Error) interface.
Not only does the interface give anonymous access to the original
[error](http://golang.org/ref/spec#Errors),
but provides the package type (e.g.,
[storage.StorageAccountsClient](https://godoc.org/github.com/Azure/azure-sdk-for-go/arm/storage#StorageAccountsClient)),
the failing method (e.g.,
[CheckNameAvailability](https://godoc.org/github.com/Azure/azure-sdk-for-go/arm/storage#StorageAccountsClient.CheckNameAvailability)),
and a detailed error message.
