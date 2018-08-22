package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec"`
	Status FunctionStatus `json:"status"`
}

type FunctionSpec struct {
	Image                  string             `json:"image"`
	Replicas               int32              `json:"replicas"`
	Class                  string             `json:"class"`
	Input                  string             `json:"input"`
	InputKeyDeserializer   string             `json:"inputKeyDeserializer"`
	InputValueDeserializer string             `json:"inputValueDeserializer"`
	Output                 string             `json:"output"`
	OutputKeySerializer    string             `json:"outputKeySerializer"`
	OutoutValueSerializer  string             `json:"outputValueSerializer"`
	FunctionConfig         *map[string]string `json:"function"`
	ConsumerConfig         *map[string]string `json:"consumer"`
	ProducerConfig         *map[string]string `json:"producer"`
}

type FunctionStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`
	AvailableReplicas  int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Function `json:"items"`
}
