package controller

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/dajac/kfn/pkg/apis/kfn/v1alpha1"
	"github.com/golang/glog"
)

// FunctionDefaultConfig contains the default configuration used for each Function.
// It is usually provided via cli's arguments
type FunctionDefaultConfig struct {
	KafkaBoostrap string
	Function      map[string]string
	Consumer      map[string]string
	Producer      map[string]string
}

// FunctionConfig represents the configuration passed to the Function runtime.
type FunctionConfig struct {
	Function map[string]string
	Consumer map[string]string
	Producer map[string]string
}

func newFunctionConfig(
	defaultConfig *FunctionDefaultConfig,
	function *v1alpha1.Function) *FunctionConfig {

	cfg := &FunctionConfig{
		Function: make(map[string]string),
		Consumer: make(map[string]string),
		Producer: make(map[string]string),
	}

	glog.Infof("%+v", defaultConfig)

	// Default config
	cfg.setKafkaBootstrapProperties(defaultConfig.KafkaBoostrap)
	cfg.overrideFunctionProperties(defaultConfig.Function)
	cfg.overrideConsumerProperties(defaultConfig.Consumer)
	cfg.overrideProducerProperties(defaultConfig.Producer)

	glog.Infof("%+v", cfg)

	// Function config
	cfg.setFunctionProperties(function)
	cfg.setSerializerDeserializer(function)

	if function.Spec.FunctionConfig != nil {
		cfg.overrideFunctionProperties(*function.Spec.FunctionConfig)
	}

	if function.Spec.ConsumerConfig != nil {
		cfg.overrideConsumerProperties(*function.Spec.ConsumerConfig)
	}

	if function.Spec.ProducerConfig != nil {
		cfg.overrideProducerProperties(*function.Spec.ProducerConfig)
	}

	return cfg
}

func (cfg *FunctionConfig) setKafkaBootstrapProperties(kafkaBootstrap string) {
	cfg.Consumer["bootstrap.servers"] = kafkaBootstrap
	cfg.Producer["bootstrap.servers"] = kafkaBootstrap
}

func (cfg *FunctionConfig) setSerializerDeserializer(function *v1alpha1.Function) {
	cfg.Consumer["key.deserializer"] = getDeserializer(function.Spec.InputKeyDeserializer)
	cfg.Consumer["value.deserializer"] = getDeserializer(function.Spec.InputValueDeserializer)

	cfg.Producer["key.serializer"] = getSerializer(function.Spec.OutputKeySerializer)
	cfg.Producer["value.serializer"] = getSerializer(function.Spec.OutoutValueSerializer)
}

func getSerializer(name string) string {
	switch name {
	case "bytes":
		return "org.apache.kafka.common.serialization.ByteArraySerializer"
	case "string":
		return "org.apache.kafka.common.serialization.StringSerializer"
	case "double":
		return "org.apache.kafka.common.serialization.DoubleSerializer"
	case "float":
		return "org.apache.kafka.common.serialization.FloatSerializer"
	case "int":
		return "org.apache.kafka.common.serialization.IntegerSerializer"
	case "long":
		return "org.apache.kafka.common.serialization.LongSerializer"
	case "short":
		return "org.apache.kafka.common.serialization.ShortSerializer"
	default:
		return name
	}
}

func getDeserializer(name string) string {
	switch name {
	case "bytes":
		return "org.apache.kafka.common.serialization.ByteArrayDeserializer"
	case "string":
		return "org.apache.kafka.common.serialization.StringDeserializer"
	case "double":
		return "org.apache.kafka.common.serialization.DoubleDeserializer"
	case "float":
		return "org.apache.kafka.common.serialization.FloatDeserializer"
	case "int":
		return "org.apache.kafka.common.serialization.IntegerDeserializer"
	case "long":
		return "org.apache.kafka.common.serialization.LongDeserializer"
	case "short":
		return "org.apache.kafka.common.serialization.ShortDeserializer"
	default:
		return name
	}
}

func (cfg *FunctionConfig) setFunctionProperties(function *v1alpha1.Function) {
	cfg.Function["name"] = function.Name
	cfg.Function["class"] = function.Spec.Class
	cfg.Function["input"] = function.Spec.Input
	cfg.Function["output"] = function.Spec.Output

	cfg.Consumer["group.id"] = function.Name
}

func (cfg *FunctionConfig) overrideFunctionProperties(src map[string]string) {
	copyWithPrefix(src, cfg.Function)
}

func (cfg *FunctionConfig) overrideConsumerProperties(src map[string]string) {
	copyWithPrefix(src, cfg.Consumer)
}

func (cfg *FunctionConfig) overrideProducerProperties(src map[string]string) {
	copyWithPrefix(src, cfg.Producer)
}

// SerializeAsProperties serializes the FunctionConfig as properties
func (cfg *FunctionConfig) SerializeAsProperties() string {
	builder := strings.Builder{}

	serializeMap(&builder, cfg.Function, "function")
	builder.WriteString("\n")
	serializeMap(&builder, cfg.Consumer, "consumer")
	builder.WriteString("\n")
	serializeMap(&builder, cfg.Producer, "producer")

	return builder.String()
}

func serializeMap(builder *strings.Builder, props map[string]string, prefix string) {
	sortedKeys := make([]string, 0, len(props))

	for key := range props {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		builder.WriteString(fmt.Sprintf("%s.%s=%s\n", prefix, key, escape(props[key])))
	}
}

func escape(s string) string {
	buffer := bytes.Buffer{}

	for _, ch := range s {
		switch ch {
		case '\n':
			buffer.WriteString("\\n")
		case '\r':
			buffer.WriteString("\\r")
		default:
			buffer.WriteRune(ch)
		}
	}

	return buffer.String()
}

func copyWithPrefix(src map[string]string, dst map[string]string) {
	if src == nil {
		return
	}

	for key, value := range src {
		dst[key] = value
	}
}
