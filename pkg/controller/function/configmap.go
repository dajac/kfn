package function

import (
	"crypto/sha256"
	"encoding/hex"

	kfnv1alpha1 "github.com/dajac/kfn/pkg/apis/kfn/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newConfigMap(function *kfnv1alpha1.Function, config *FunctionConfig) *corev1.ConfigMap {
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
			"function.properties": config.SerializeAsProperties(),
		},
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
