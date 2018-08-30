# Getting Started with KFn Function Deployement

## Prerequisites

Before you begin, you need:

- a Kubernetes cluster with [Kafka and KFn installed](https://github.com/dajac/kfn/blob/master/docs/install-with-any-k8s.md).

## Sample Function

This guide uses the [CopyFunction sample Function in Java](https://github.com/dajac/kfn-examples/blob/master/src/main/java/io/dajac/kfn/examples/CopyFunction.java) to demonstrate the basic workflow for deploying a Function, but these steps can be adapted for your own Function if you have an image available on Docker Hub, or another container image registry.

The `CopyFunction` reads messages from a given topic and writes them to another one.

```java
package io.dajac.kfn.examples;

import io.dajac.kfn.invoker.Function;
import io.dajac.kfn.invoker.KeyValue;

public class CopyFunction implements Function<byte[], byte[], byte[], byte[]> {
    @Override
    public KeyValue<byte[], byte[]> apply(byte[] key, byte[] value) {
        return KeyValue.pair(key, value);
    }
}
```

### Configuring your Function

To deploy a Function using KFn, you need to create a manifest that defines the Function. For more information about the Function object, see the [type definition](https://github.com/dajac/kfn/blob/master/pkg/apis/kfn/v1alpha1/types.go).

Create a new file named `copy-function.yaml`, the copy and paste the following content onto it:

```yaml
apiVersion: kfn.dajac.io/v1alpha1
kind: Function
metadata:
  name: copy-function
spec:
  replicas: 1
  image: dajac/kfn-examples:0.1.0
  class: io.dajac.kfn.examples.CopyFunction
  
  input: kfn.source
  inputKeyDeserializer: bytes
  inputValueDeserializer: bytes
  
  output: kfn.destination
  outputKeySerializer: bytes
  outputValueSerializer: bytes
```

## Creating the topics

Before deploying your Function, you need to create the topics that it uses.

1. Run `kubectl apply` command to deploy a kafka client:

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/dajac/kfn/master/docs/install-with-any-k8s/kafka-client.yaml
    ```

2. Use `kubectl exec` command to open a shell within the container:

    ```bash
    kubectl exec -ti kafka-client bin/sh
    ```

3. Create the topics with the `kafka-topics` command:

    ```bash
    /usr/bin/kafka-topics --zookeeper zookeeper:2181 --create --topic kfn.source --partitions 5 --replication-factor 1
    /usr/bin/kafka-topics --zookeeper zookeeper:2181 --create --topic kfn.destination --partitions 5 --replication-factor 1
    ```

## Deploying your Function

From the directory where the `copy-function.yaml` file was created, apply the manifest:

```bash
kubectl apply -f copy-function.yaml
```

Now that your Function is created, the KFn operator will deploy it.

## Interacting with your Function

To see if your Function has been deployed successfuly, you can see its status with the following command:

```bash
kubectl get functions
```

You will see the number of desired replicas and the number of replicas that are actually running. It can take a while to pull the image of the container.

## Validating the deployement

To validate the deployement, let's produce few messages to the intut topic of your Function and consume the messages from the output topic.

1. Run `kubectl apply` command to deploy a kafka client:

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/dajac/kfn/master/docs/install-with-any-k8s/kafka-client.yaml
    ```

2. Use `kubectl exec` command to open a shell within the container:

    ```bash
    kubectl exec -ti kafka-client bin/sh
    ```

3. Produce few messages with the `kafka-console-producer` command:

    ```bash
    /usr/bin/kafka-console-producer --broker-list kafka-headless:9092 --topic kfn.source
    ```

4. Consume the messages with the `kafka-console-consumer` command:

    ```bash
    /usr/bin/kafka-console-consumer --bootstrap-server kafka-headless:9092 --topic kfn.destination --from-beginning
    ```

## Scaling up and down your Function

`kubectl scale` command can be used to scale up and down your Function like you would do for a deployement:

```bash
kubectl scale --replicas=2 function/copy-function
```

## Cleaning up

To remove the sample Function, delete the function with the following command:

```bash
kubectl delete function copy-function
```
