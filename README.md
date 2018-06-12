# Kong Operator

[![Build Status](https://travis-ci.org/Cdiscount/kong-operator.svg?branch=master)](https://travis-ci.org/Cdiscount/kong-operator)
[![codecov](https://codecov.io/gh/Cdiscount/kong-operator/branch/master/graph/badge.svg)](https://codecov.io/gh/Cdiscount/kong-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/Cdiscount/kong-operator)](https://goreportcard.com/report/github.com/Cdiscount/kong-operator)

**Project status: *beta*** Not all planned features are completed. This operator aim to manage Kong *routes* and *services* to expose Kubernetes services.

Once installed the Kong Operator provides the following features:
* **Create/Destroy/Update** Kong services using all parameters offer by Kong API.
* **Target Kubernetes Services** Automatically expose Kubernetes services through labels.

## Prerequistes
Kong operator supported:
* Kubernetes version `>= 1.8`
* Kong version `>= 0.13.0`

## Installation

```shell
kubectl apply -f manifests/
```

This command will create:
* A `namespace` `kong-operator`
* All `RBAC` ressources needed
* A `deployment` to manage `kong-operator` pod

## CustomResourceDefinition

Kong Operator acts on the following [custom resource definitions (CRDs)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/):
* `KongRoute`: this CRD define a new [Kong route](https://getkong.org/docs/0.13.x/admin-api/#route-object) to expose a `KongService`
* `KongService`: wich defined a [Kong service](https://getkong.org/docs/0.13.x/admin-api/#service-object). It will represent a `kubernetes service`

For more informations, check the [design doc](docs/design.md)
