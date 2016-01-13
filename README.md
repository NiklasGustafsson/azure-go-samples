# Azure Go SDK Samples

This repository contains samples illustrating how to use the Go SDK for Azure.

The Go SDK is experimental and offers only partial coverage what is available in Azure. Just like the SDK is divided into
three high-level areas of functionality, the samples are organized in three collections:

[Azure Storage](./storage/blobs)

These samples cover the Azure Storage APIs for blobs, queues, and files. They are fairly simple, explaining how to utilize the
basic features.

[Service Management](./service_management)

The service management are the traditional Azure API to manage services such as virtual machines, networks, and databases.
Typically, these would be called from tools and process automation logic, not applications.

[Azure Resource Manager](./arm)

The Azure Resource Manager replaces the Service Management APIs with a more composable and extensible framework for describing
computational and data resources in Azure. It comes with some cost in terms of complexity and a change of authentication models,
but once you're past the learning curve, ARM is very powerful.

