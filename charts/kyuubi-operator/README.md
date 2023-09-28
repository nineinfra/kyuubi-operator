# Helm Chart for Kyuubi Operator for Apache Kyuubi

This Helm Chart can be used to install Custom Resource Definitions and the Operator for Apache Kyuubi provided by Nineinfra.

## Requirements

- Create a [Kubernetes Cluster](../Readme.md)
- Install [Helm](https://helm.sh/docs/intro/install/)

## Install the Kyuubi Operator for Apache Kyuubi

```bash
# From the root of the operator repository

helm install kyuubi-operator charts/kyuubi-operator
```

## Usage of the CRDs

The usage of this operator and its CRDs is described in the [documentation](https://github.com/nineinfra/kyuubi-operator/blob/main/README.md).

## Links

https://github.com/nineinfra/kyuubi-operator
