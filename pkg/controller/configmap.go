package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	kfnv1alpha1 "github.com/dajac/kfn/pkg/apis/kfn/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

/* Example
apiVersion: v1
kind: ConfigMap
metadata:
  name: kfn-copy-function-config
  namespace: default
data:
  function.properties: |-
    consumer.bootstrap.servers=kafka-headless:9092
    consumer.group.id=copy-function
    consumer.auto.offset.reset=earliest
    consumer.key.deserializer=org.apache.kafka.common.serialization.ByteArrayDeserializer
    consumer.value.deserializer=org.apache.kafka.common.serialization.ByteArrayDeserializer

    producer.bootstrap.servers=kafka-headless:9092
    producer.acks=all
    producer.retries=3
    producer.key.serializer=org.apache.kafka.common.serialization.ByteArraySerializer
    producer.value.serializer=org.apache.kafka.common.serialization.ByteArraySerializer

    function.name=copy-function
    function.class=io.dajac.kfn.functions.CopyFunction
    function.input=dajac.dev.input
	function.output=dajac.dev.output
*/

var defaultProperties = map[string]string{
	"consumer.bootstrap.servers":  "kafka-headless:9092",
	"consumer.group.id":           "CHANGE ME",
	"consumer.auto.offset.reset":  "earliest",
	"consumer.key.deserializer":   "org.apache.kafka.common.serialization.ByteArrayDeserializer",
	"consumer.value.deserializer": "org.apache.kafka.common.serialization.ByteArrayDeserializer",

	"producer.bootstrap.servers": "kafka-headless:9092",
	"producer.acks":              "all",
	"producer.retries":           "3",
	"producer.key.serializer":    "org.apache.kafka.common.serialization.ByteArraySerializer",
	"producer.value.serializer":  "org.apache.kafka.common.serialization.ByteArraySerializer",
}

func newConfigMap(function *kfnv1alpha1.Function) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      function.Name,
			Namespace: function.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(function, schema.GroupVersionKind{
					Group:   kfnv1alpha1.SchemeGroupVersion.Group,
					Version: kfnv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Function",
				}),
			},
		},
		Data: map[string]string{
			"function.properties": serializeProperties(getProperties(function)),
		},
	}
}

func serializeProperties(props map[string]string) string {
	sortedKeys := make([]string, 0, len(props))

	for key := range props {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys)

	builder := strings.Builder{}

	for _, key := range sortedKeys {
		builder.WriteString(fmt.Sprintf("%s=%s\n", key, props[key]))
	}

	return builder.String()
}

func getProperties(function *kfnv1alpha1.Function) map[string]string {
	props := map[string]string{
		"function.name":   function.Name,
		"function.class":  function.Spec.Class,
		"function.input":  function.Spec.Input,
		"function.output": function.Spec.Output,

		"consumer.group.id": function.Name,
	}

	if function.Spec.FunctionConfig != nil {
		copyWithPrefix(*function.Spec.FunctionConfig, props, "function")
	}

	if function.Spec.ConsumerConfig != nil {
		copyWithPrefix(*function.Spec.ConsumerConfig, props, "consumer")
	}

	if function.Spec.ProducerConfig != nil {
		copyWithPrefix(*function.Spec.ProducerConfig, props, "producer")
	}

	return mergeMap(defaultProperties, props)
}

func copyWithPrefix(src map[string]string, dst map[string]string, prefix string) {
	for key, value := range src {
		dst[fmt.Sprintf("%s.%s", prefix, key)] = value
	}
}

func mergeMap(a map[string]string, b map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range a {
		result[key] = value
	}

	for key, value := range b {
		result[key] = value
	}

	return result
}

func hash(configMap *corev1.ConfigMap) string {
	props := configMap.Data["function.properties"]
	h := sha256.New()
	h.Write([]byte(props))
	return hex.EncodeToString(h.Sum(nil))
}
