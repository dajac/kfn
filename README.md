# KFn - Serverless Function made easy for Apache Kafka

KFn (Kafka Function) is a framework for building serverless Functions for Apache Kafka. It lets you deploy functions written in Java or any JVM based languages, without having to worry about the underlying infrastructure. KFn leverages Kubernetes ressources to make it happen.

**It is an alpha software, don't use in production!**

## Highligths

* Support any JVM based languages (Java, Clojure, Scala, etc.), and package in Docker image
* Use the native Kafka Java consumer and producer, no stdin/stdout nor RPC between the Function and Kafka
* Use Custom Ressource Definition in Kubernetes
* Automatic rolling restart when the config changes
* Autoscaling (still in development)
* Should run on any Kubernetes cluster (v1.11 or newer)

## Documentations

* [Installing KFn](https://github.com/dajac/kfn/blob/master/docs/install-with-any-k8s.md)
* [Getting Started with KFn Function Deployement](https://github.com/dajac/kfn/blob/master/docs/getting-started.md)
* [Deploy advanced KFn Functions](https://github.com/dajac/kfn/blob/master/docs/advanced-example.md)

## Dependencies

* [kfn-invoker](https://github.com/dajac/kfn-invoker) - The KFn invoker used in the container to invoke the Function
* [kfn-examples](https://github.com/dajac/kfn-examples) - Examples

## Known Limitations

* It will be possible to connect to secure Kafka cluster only when secrets will be supported by KFn.
