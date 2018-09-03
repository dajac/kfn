# Deploy advanced KFn Functions

## Prerequisites

Before you begin, you need:

- a Kubernetes cluster with [Kafka and KFn installed](https://github.com/dajac/kfn/blob/master/docs/install-with-any-k8s.md)
- understanding of the worklow to [deploy a Function](https://github.com/dajac/kfn/blob/master/docs/getting-started.md)

## Multiple Functions

This guide uses the [JsonToAvroFunction](https://github.com/dajac/kfn-examples/blob/master/src/main/java/io/dajac/kfn/examples/JsonToAvroFunction.java) and [HashFieldFunction](https://github.com/dajac/kfn-examples/blob/master/src/main/java/io/dajac/kfn/examples/HashFieldFunction.java) to demonstrate the deployement of two Functions. Let's assume that your have a topic containing message serialised in JSON format. You would like to transform that topic to have it serialised in Avro format and you would like to have another topic where one of the field is hashed to anonymise the data.

## Configuring your Functions

Create a new file name `functions.yaml`, then copy and paste the following content into it:

```yaml
apiVersion: kfn.dajac.io/v1alpha1
kind: Function
metadata:
  name: json-avro-converter-function
spec:
  replicas: 1
  image: dajac/kfn-examples:0.1.0
  class: io.dajac.kfn.examples.JsonToAvroFunction
  
  input: kfn.users.json
  inputKeyDeserializer: bytes
  inputValueDeserializer: string
  
  output: kfn.users.avro
  outputKeySerializer: string
  outputValueSerializer: io.confluent.kafka.serializers.KafkaAvroSerializer

  function:
    key: name
    schema: |- 
      {
        "type": "record",
        "name": "User",
        "namespace": "example",
        "fields": [
          {"name": "name", "type": "string"},
          {"name": "number", "type": "string"}
        ]
      }
  
  producer:
    schema.registry.url: "http://schema-registry-service:8081"
---
apiVersion: kfn.dajac.io/v1alpha1
kind: Function
metadata:
  name: hash-field-function
spec:
  replicas: 1
  image: dajac/kfn-examples:0.1.0
  class: io.dajac.kfn.examples.HashFieldFunction
  
  input: kfn.users.avro
  inputKeyDeserializer: string
  inputValueDeserializer: io.confluent.kafka.serializers.KafkaAvroDeserializer
  
  output: kfn.users.avro.hashed
  outputKeySerializer: string
  outputValueSerializer: io.confluent.kafka.serializers.KafkaAvroSerializer

  function:
    field: number
    algorythm: SHA-256

  consumer:
    schema.registry.url: "http://schema-registry-service:8081"

  producer:
    schema.registry.url: "http://schema-registry-service:8081"
```

The first Function `json-avro-converter-function` reads from the topic `kfn.users.json` and write to `kfn.users.avro`. It uses the schema `function.schema` to parse the JSON documents and serialise them in Avro and it uses `function.name` to extract the key that will be used while producing messages.

The second Function `hash-field-function` reads from the topic `kfn.users.avro` and write to `kfn.users.avro.hashed`. It also passes specific configurations to the underlying consumer and producer to pass the URL of the schema registry.

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
    /usr/bin/kafka-topics --zookeeper zookeeper:2181 --create --topic kfn.users.json --partitions 5 --replication-factor 1
    /usr/bin/kafka-topics --zookeeper zookeeper:2181 --create --topic kfn.users.avro --partitions 5 --replication-factor 1
    /usr/bin/kafka-topics --zookeeper zookeeper:2181 --create --topic kfn.users.avro.hashed --partitions 5 --replication-factor 1
    ```

## Deploying your Functions

From the directory where the `functions.yaml` file was created, apply the manifest:

```bash
kubectl apply -f functions.yaml
```

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

3. Produce few messages with the `kafka-console-producer` command. You can use the following JSON document as an example: `{"name": "David", "number": "+41791234567"}`

    ```bash
    /usr/bin/kafka-console-producer --broker-list kafka-headless:9092 --topic kfn.users.json
    ```

4. Consume the messages with the `kafka-console-consumer` command:

    ```bash
    /usr/bin/kafka-console-consumer --bootstrap-server kafka-headless:9092 --topic kfn.users.avro.hashed --from-beginning
    ```

## Cleaning up

To remove the sample Function, delete the function with the following command:

```bash
kubectl delete function json-avro-converter-function
kubectl delete function hash-field-function
```
