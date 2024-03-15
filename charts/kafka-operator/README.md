# Helm Chart for Hdfs Operator for Apache Kafka

This Helm Chart can be used to install Custom Resource Definitions and the Operator for Apache Kafka provided by Nineinfra.

## Requirements

- Create a [Kubernetes Cluster](../Readme.md)
- Install [Helm](https://helm.sh/docs/intro/install/)

## Install the Hdfs Operator for Apache Kafka

```bash
# From the root of the operator repository

helm install kafka-operator charts/kafka-operator
```

## Usage of the CRDs

The usage of this operator and its CRDs is described in the [documentation](https://github.com/nineinfra/kafka-operator/blob/main/README.md).

## Links

https://github.com/nineinfra/kafka-operator
