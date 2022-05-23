# Accessing CRDs from the client-go package
The Kubernetes API server is easily extendable by Custom Resource Definitions. However, actually accessing these resources from the popular client-go library is a bit more complex and not thoroughly documented. This article contains a short guide on how to access CRDs from your own Go code.

## Motivation
I came across this challenge while my customer wanting to integrate a Kasten functions into a Kubernetes cluster Management Platform. The plan was to use CRDs to define things like Backup Policies and Location Profiles.

## Project Brief
In my code, we’ll work with an easy example to show how to Accessing  Kasten Profiles CRD like your daily work on kubernetes resources. I have wrote the following functions for a example, enjoy and feel free to let me know want the problem or concern you care about. Enjoy :-) Mars

## Functions Covered
 - GET
 - LIST
 - CREATE
 - PATCH
 - DELETE
 - UPDATE
