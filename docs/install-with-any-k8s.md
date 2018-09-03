# KFn install on a Kubernetes cluster

This guide walks you through the installation of the latest version of KFn using a pre-built image.

## Before you begin

KFn requires a Kubernetes cluster v1.11 or newer. `kubectl` v1.11 is also required. This guide assumes that you've created a Kubernetes cluster which you're comfortable installing alpha software on. This guide assumes that you're using bash in a Mac or Linux environment.

## Installing Zookeeper, Kafka and Schema Registry

KFn Functions will consume from and produce to a Kafka cluster. The Schema Registry is also required for one of the exemple Function. If you already have those running in your Kubernetes cluster, you can safely skip this chapter. Otherwise, we provide a basic setup that you can install with the following instructions.

1. Run the `kubectl apply` commands to install all the components in the `default` namespace:

```bash
kubectl apply -f https://raw.githubusercontent.com/dajac/kfn/master/docs/install-with-any-k8s/zookeeper.yaml
kubectl apply -f https://raw.githubusercontent.com/dajac/kfn/master/docs/install-with-any-k8s/kafka.yaml
kubectl apply -f https://raw.githubusercontent.com/dajac/kfn/master/docs/install-with-any-k8s/schema-registry.yaml
```

2. Monitor the components until all components show a `STATUS` of `Running`:

```bash
kubectl get pods
```

## Installing KFn

KFn can be installed by using the manifest provided in the [releases](https://github.com/dajac/kfn/releases). If you use our default Kafka setup, you can directly deploy the manifest. Otherwise, you have to adapt the Kafka's FQDN in the manifest to target your cluster.

1. Run the `kubectl apply` command to install the KFn operator in the `kfn` namespace":

```bash
kubectl apply -f https://github.com/dajac/kfn/releases/download/v0.1.0/kfn-0.1.0.yaml
```

2. Monitor the operator until its shows a `STATUS` of `Running`:

```bash
kubectl get pods -n kfn
```

3. Query the Functions with:

```bash
kubectl get functions
```

## Deploying a Function

Now that your cluster has KFn installed, you're ready to deploy a Function. You can follow the step-by-step [Getting Started](https://github.com/dajac/kfn/blob/master/docs/getting-started.md) guide.
