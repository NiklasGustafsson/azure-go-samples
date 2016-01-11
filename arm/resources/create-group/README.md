# Creating a Resource Group

One of the most fundamental concepts of the Azure Resource Manager (ARM) is that of a *resource group*. Any Azure resource exists within a
resource group, which determines its location. Thus, before we can create any resources, we must create the resource group within which
we will place the resources.

In this example, which presupposes that you have taken a look at the 'check-name' sample under 'arm/auth' to get an understanding of how
ARM authenticates clients. The authentication logic in that sample has been added to the 'helpers' package, since we don't want to clutter
up our logic with generic auth logic. Note, however, that the AuthenticateForARM() function is not an SDK API, it is just a helper
for use in the samples, since it relies on unsecure storage of credentials in plain files.

## The Code

The code to programmatically create a resource group is fairly simple and straight-forward. As preparatipn, we authenticate, 
use the results to create a resource groups client, attach the authentication logic to it, and add some debug loggers.
```go
	groupName := "armtestgroup"
	groupLocation := "West US"

	spt,sid,err := helpers.AuthenticateForARM()
	
	rgc := resources.NewGroupsClient(sid)
	rgc.Authorizer = spt	
	
	rgc.RequestInspector = helpers.WithInspection()
	rgc.ResponseInspector = helpers.ByInspecting()
```
Then, we are ready to call the resource manager directly. We need to pass a struct with the name and location of the
resource group to CreateOrUpdate(), an idempotent factory API. That's it!
Use the [new Azure portal](http://portal.azure.com) to verify that a resource group named 'armtestgroup' was indeed created after you run the code
sample.  
```go
	params := resources.ResourceGroup{Name:&groupName, Location:&groupLocation}
	
	group,err := rgc.CreateOrUpdate(groupName, params)
	if err != nil {
		fmt.Printf("Failed to create resource group '%s' in location '%s': '%s'\n", groupName, groupLocation, err.Error())
		return
	}
```