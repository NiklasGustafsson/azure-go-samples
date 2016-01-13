# Azure Resource Manager

The samples in this section build on each other, so in order to get the most out of them, we suggest that you
look at them in the following order:

[Authenticiation](./auth/check-name)

ARM authentication is, unfortunately, relatively complex. In fact, the most complex part of it all is to set up your account to
accommodate the fact that ARM expects the client to have an application identity, a *service principal*. How to do this is explained
in the sample commentary, and you will not be able to run any of the other samples without first getting that setup taken care of,
so start here.

[Resource Groups](./resources/create-group)

The most basic notion in ARM is that of a resource group, which are collections of resources within your subscription. Before you
create any resources, you must create a group to contain them, no resource ever exists outside a group. This sample demonstrates
how this is done.

[Virtual Machine Creation](./resources/create-vm)

Similar in nature to one of the samples under the Service Management section, this sample will allow you to programmatically create
a single Windows-based virtual machine in Azure and then connect to it using Remote Desktop. It's fairly complex, involving several
steps to create, then combine, the building blocks needed to set up a VM that you can communicate with.

[Azure Resource Manager Templates](./templates/deploy-template)

One of the more powerful features of ARM is the ability to separate the definition of a collection of resources from the logic that
creates them. ARM templates are documents that are sent to Azure and associated with a resource group in an operation called a deployment.
The act of realizing the deployment is then finished within Azure, thus freeing the client from the need to orchestrate resource
creation, handle errors, and logging the events.