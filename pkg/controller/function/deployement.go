package function

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kfnv1alpha1 "github.com/dajac/kfn/pkg/apis/kfn/v1alpha1"
)

/* Example
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kfn-copy-function
  labels:
    app: kfn-copy-function
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kfn-copy-function
  template:
    metadata:
      labels:
        app: kfn-copy-function
    spec:
      containers:
      - name: kfn-invoker
        image: dajac/kfn-invoker:0.1.0
        imagePullPolicy: Always
        command:
          - /usr/bin/java
          - -cp
          - /usr/lib/kfn/*
          - io.dajac.kfn.invoker.FunctionInvoker
          - /etc/kfn/function.properties
        volumeMounts:
        - name: config-vol
          mountPath: /etc/kfn
      volumes:
      - name: config-vol
        configMap:
          name: kfn-copy-function-config
          items:
          - key: function.properties
			path: function.properties
*/

func newDeployement(function *kfnv1alpha1.Function, configMap *corev1.ConfigMap) *appsv1.Deployment {
	labels := map[string]string{
		"function": function.Name,
	}

	return &appsv1.Deployment{
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
		Spec: appsv1.DeploymentSpec{
			Replicas: &function.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"kfn.dajac.io/config-hash": hash(configMap),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "kfn-invoker",
							Image:           function.Spec.Image,
							ImagePullPolicy: "Always",
							Command: []string{
								"/usr/bin/java",
								"-cp",
								"/usr/lib/kfn/*",
								"io.dajac.kfn.invoker.FunctionInvoker",
								"/etc/kfn/function.properties",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "configuration",
									MountPath: "/etc/kfn",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "configuration",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: function.Name,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "function.properties",
											Path: "function.properties",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
