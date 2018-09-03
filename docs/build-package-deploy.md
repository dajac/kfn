# Build, Package and Deploy your own KFn Function

## Prerequisites

Before you begin, you need:

- a Kubernetes cluster with [Kafka and KFn installed](https://github.com/dajac/kfn/blob/master/docs/install-with-any-k8s.md)
- understanding of the worklow to [deploy a Function](https://github.com/dajac/kfn/blob/master/docs/getting-started.md)
- Maven and Docker

## Build

To build a Function with KFn, you can use the following instructions. Instructions are based on this [template](https://github.com/dajac/kfn-template).

1. Create a `pom.xml` that includes the kfn-invoker as a dependency:

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>

        <groupId>io.dajac.kfn</groupId>
        <artifactId>kfn-myfunction</artifactId>
        <version>0.1.0</version>

        <dependencies>
            <!-- kfn-invoker -->
            <dependency>
                <groupId>io.dajac.kfn</groupId>
                <artifactId>kfn-invoker</artifactId>
                <version>0.1.0</version>
                <scope>provided</scope>
            </dependency>
            <dependency>
                <groupId>junit</groupId>
                <artifactId>junit</artifactId>
                <version>4.12</version>
                <scope>test</scope>
            </dependency>
        </dependencies>

        <repositories>
            <!-- repository where kfn-invoker is -->
            <repository>
                <id>clojars</id>
                <url>http://clojars.org/repo/</url>
            </repository>
        </repositories>

        <build>
            <plugins>
                <plugin>
                    <groupId>org.apache.maven.plugins</groupId>
                    <artifactId>maven-compiler-plugin</artifactId>
                    <version>3.1</version>
                    <configuration>
                        <source>1.8</source>
                        <target>1.8</target>
                    </configuration>
                </plugin>
                <!-- copy all deps into target/libs -->
                <plugin>
                    <artifactId>maven-dependency-plugin</artifactId>
                    <version>3.0.1</version>
                    <executions>
                        <execution>
                            <id>copy-dependencies</id>
                            <phase>package</phase>
                            <goals>
                                <goal>copy-dependencies</goal>
                            </goals>
                            <configuration>
                                <outputDirectory>${project.build.directory}/libs</outputDirectory>
                                <prependGroupId>true</prependGroupId>
                                <excludeScope>provided</excludeScope>
                            </configuration>
                        </execution>
                    </executions>
                </plugin>
            </plugins>
        </build>
    </project>

    ```

2. Write a KFn Function in `src/main/java/MyFunction.java`:

    ```java
    import io.dajac.kfn.invoker.Function;
    import io.dajac.kfn.invoker.KeyValue;

    public class MyFunction implements Function<byte[], byte[], byte[], byte[]> {
        @Override
        public KeyValue<byte[], byte[]> apply(byte[] key, byte[] value) {
            // Do something here!
            return KeyValue.pair(key, value);
        }
    }
    ```

3. Build the project:

    ```bash
    mvn package
    ```

## Package

KFn relies on Docker to package and ship the Functions. You can package your KFn Function with the following instructions:

1. Create a `Dockerfile` that uses the KFn base image `dajac/kfn-invoker:0.1.0`, includes all the dependencies and the applications in `/usr/lib/kfn/`:

    ```dockerfile
    FROM dajac/kfn-invoker:0.1.0
    COPY target/libs/* /usr/lib/kfn/
    COPY target/kfn-myfunction-0.1.0.jar /usr/lib/kfn/kfn-myfunction-0.1.0.jar
    ```

2. Build and push the Docker image:

    ```bash
    docker build -t <org>/kfn-my-function:0.1.0 .
    docker push <org>/kfn-my-function:0.1.0
    ```

## Deploy

To deploy a Function using KFn, you need to create a manifest that defines the Function. For more information about the Function object, see the [type definition](https://github.com/dajac/kfn/blob/master/pkg/apis/kfn/v1alpha1/types.go).

1. Create a new file named `my-function.yaml`, the copy and paste the following content onto it. Note that the deserializers and serialisers must match the types defined in your Function.

    ```yaml
    apiVersion: kfn.dajac.io/v1alpha1
    kind: Function
    metadata:
        name: my-function
    spec:
        replicas: 1
        image: <org>/kfn-my-function:0.1.0
        class: MyFunction
        
        input: kfn.source
        inputKeyDeserializer: bytes
        inputValueDeserializer: bytes
        
        output: kfn.destination
        outputKeySerializer: bytes
        outputValueSerializer: bytes
    ```

2. From the directory where the `copy-function.yaml` file was created, apply the manifest:

    ```bash
    kubectl apply -f copy-function.yaml
    ```

    Now that your Function is created, the KFn operator will deploy it.
