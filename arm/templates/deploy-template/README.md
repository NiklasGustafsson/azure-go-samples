# Using ARM Templates to Create Resources

In this sample, we'll examine another way to create Azure resources, with less code and more of the definition logic
contained in templates, which are documents containing the desired resource configuration. If you haven't gone through
the samples on authentication, group creation, and VM creation using ARM, you should do so first to get a sense of
what the strictly programmatic way to manipulate resources looks like.

## Templates

Typically, ARM templates consist of a description of resources that should be created within a resource group. A template
can be defined to take a set of parameters, allowing the definition of resource structure to be separated from some of
the concrete values that are desired, allowing templates to be reused. For example, the DNS name of a virtual machine may
be something that should be customized for each deployment.

A simple template defining a single Linux VM looks something like 
[this](https://raw.githubusercontent.com/Azure/azure-quickstart-templates/master/101-vm-simple-linux/azuredeploy.json). There's
a section for parameters, a section for variables (similar to declaring constants in a program header), and a resources
section.

The Go SDK allows you to provide the template in two ways -- a) by referring to a publicly available document online, such as a
GitHub-based file or Azure storage blob, or b) by loading the template as a map in memory and passing it to the deployment
API. The sample code shows how to do both of those.

## Parameters

For each parameter declared in the parameters section that does not have a default, a value must be provide at runtime. An
example parameter document is [this](https://raw.githubusercontent.com/Azure/azure-quickstart-templates/master/101-vm-simple-linux/azuredeploy.parameters.json)
In parallel fashion with the template document, the parameter document can be supplied as a link to a document in a public
location, or by loading the template as a map in memory and passing it in.

A surprising detail here, is that the Go SDK expects the parameter document format to be slightly different in the two
cases -- when passing in memory, only the value of the 'parameters' element should be passed.
  
## The Code

Let's look at the code!

We start with a nicety - a usage message. We expect nothing from the user, in which case a hard-coded parameter map and
template link will be used. If you want to customize the parameters without editing the code, then pass in the path of a file
where the parameters are found. If you do, you can also customize the template that is used by passing in a template file
path.
```go
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		fmt.Println("usage: deploy [parameter-file-name [template-file-name]]")
		return
	}
	
	deploymentName := "simplelinux"
	groupName := "templatetests"
	groupLocation := "West US"
```

After that, we try to figure out what was passed on the command-line. If a parameter file was passed in, we try to divine whether
it already has the 'parameters' element pulled out, or whether it is embedded in a document of the format shown earlier. If no parameters
file was passed in, the parameters map is created directly.

**Note**: you have to edit the admin password and DNS label in the code if you plan to run the sample. The same goes for the
`template01.parameters.json` document.

```go
	if parameterLink != nil {
		parameters,err = readMap(*parameterLink)
		if err != nil {
			fmt.Printf("Failed to read parameter file '%s': '%s'\n", *parameterLink, err.Error())
			return
		}
		if p,ok := parameters["parameters"]; ok {
			parameters = p.(map[string]interface{})
		}
	} else {
		parameters = map[string]interface{} {
			"adminUsername": makeStringParameterValue("tmpltest"),
			"adminPassword": makeStringParameterValue("<<PLEASE EDIT>>"),
			"dnsLabelPrefix": makeStringParameterValue("<<MUST BE UNIQUE>>"),
			"ubuntuOSVersion": makeStringParameterValue("14.04.2-LTS"),
		}
	}
```
Once we have clarity on the parameters, we do something similar to the template itself. If nothing was passed, we
go for the template document checked into the sample repo on Github.

```go
	var deploymentProps resources.DeploymentProperties
	
	if templateLink != nil {

		template,err := readMap(*templateLink)
		if err != nil {
			fmt.Printf("Failed to read template file '%s': '%s'\n", *templateLink, err.Error())
			return
		}

		deploymentProps = resources.DeploymentProperties {
			Template: &template,
			Parameters: &parameters,
			Mode: resources.Incremental,
		}

	} else {

		deploymentProps = resources.DeploymentProperties {
			TemplateLink: &resources.TemplateLink { 
				URI: to.StringPtr("https://raw.githubusercontent.com/.../template01.json"),
				ContentVersion: to.StringPtr("1.0.0.0"),
			},
			Parameters: &parameters,
			Mode: resources.Incremental,
		}

	}
```

After that, it's smooth sailing. The code should be very familiar to you, having gone through the other ARM samples:
```go
	depc := resources.NewDeploymentsClient(sid)
	depc.Authorizer = spt	
	
	deployment,err := depc.CreateOrUpdate(
	    groupName, 
	    deploymentName, 
	    resources.Deployment { 
	    	Properties: &deploymentProps,
	    })
```

That's it! With the exception of the hard-coded parameters map, this code is independent of what kind of template you are using,
any ARM resources may be deployed using this approach.

A great number of sample templates are available [here](https://github.com/Azure/azure-quickstart-templates). It's well worth spending
a few hours looking through the templates to get an idea of how to structure your own templates.
