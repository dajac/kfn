package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Function describes an KFn Function
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec"`
	Status FunctionStatus `json:"status"`
}

// FunctionSpec is the specification of a KFn Function ressource
type FunctionSpec struct {
	// Image is the Docker image of the Function.
	// Image must be based on dajac/kfn-invoker:x.x.x
	Image string `json:"image"`

	// Replicas is the expected number of Function.
	Replicas int32 `json:"replicas"`

	// Class is the fully qualified class name of the Function.
	Class string `json:"class"`

	// Input is the name of the input topic.
	Input string `json:"input"`

	// InputKeyDeserializer is the name of the deserializer used by
	// the Kafka Consumer to deserialize the key of each messages.
	// It accepts the following types - bytes, string, double, float,
	// int, long, short - or the fully qualified class name
	// of the Deserializer. Deserializer must be present in the image.
	// The type must match the type accepted by the Function.
	InputKeyDeserializer string `json:"inputKeyDeserializer"`

	// InputValueDeserializer is the name of the deserializer used by
	// the Kafka Consumer to deserialize the value of each messages.
	// See InputKeyDeserializer for details.
	InputValueDeserializer string `json:"inputValueDeserializer"`

	// Output is the name of the output topic.
	Output string `json:"output"`

	// OutputKeySerializer is the name of the serializer used by
	// the Kafka Producer to serialize the key of each messages.
	// See InputKeyDeserializer for details.
	OutputKeySerializer string `json:"outputKeySerializer"`

	// OutoutValueSerializer is the name of the serializer used by
	// the Kafka Producer to serialize the value of each messages.
	// See InputKeyDeserializer for details.
	OutoutValueSerializer string `json:"outputValueSerializer"`

	// FunctionConfig is a set of key-value pairs which will be passed to
	// the Function via the `configure` method.
	FunctionConfig *map[string]string `json:"function"`

	// ConsumerConfig is a set of key-value pairs which will be passed to
	// the Kafka Consumer.
	ConsumerConfig *map[string]string `json:"consumer"`

	// ProducerConfig is a set of key-value pairs which will be passed to
	// the Kafka Producer.
	ProducerConfig *map[string]string `json:"producer"`
}

// FunctionStatus describes the status of a KFn Function
type FunctionStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`
	AvailableReplicas  int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FunctionList is a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Function `json:"items"`
}
