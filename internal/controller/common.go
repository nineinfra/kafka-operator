package controller

import (
	"fmt"
	kafkav1 "github.com/nineinfra/kafka-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func ClusterResourceName(cluster *kafkav1.KafkaCluster, suffixs ...string) string {
	return cluster.Name + DefaultNameSuffix + strings.Join(suffixs, "-")
}

func ClusterResourceLabels(cluster *kafkav1.KafkaCluster) map[string]string {
	return map[string]string{
		"cluster": cluster.Name,
		"app":     DefaultClusterSign,
	}
}

func GetStorageClassName(cluster *kafkav1.KafkaCluster) string {
	if cluster.Spec.Resource.StorageClass != "" {
		return cluster.Spec.Resource.StorageClass
	}
	return DefaultStorageClass
}

func GetFullSvcName(cluster *kafkav1.KafkaCluster) string {
	return fmt.Sprintf("%s.%s.svc.%s", ClusterResourceName(cluster), cluster.Namespace, GetClusterDomain(cluster))
}

func GetClusterDomain(cluster *kafkav1.KafkaCluster) string {
	if cluster.Spec.K8sConf != nil {
		if value, ok := cluster.Spec.K8sConf[DefaultClusterDomainName]; ok {
			return value
		}
	}
	return DefaultClusterDomain
}

func DefaultDownwardAPI() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_UID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.uid",
				},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		},
	}
}

func DefaultEnvs() []corev1.EnvVar {
	envs := DefaultDownwardAPI()
	envs = append(envs,
		corev1.EnvVar{
			Name:  "KAFKA_LOG4J_OPTS",
			Value: fmt.Sprintf("-Dlog4j.configuration=file:%s/%s", DefaultConfPath, DefaultLogConfigFileName),
		},
		corev1.EnvVar{
			Name:  "INTERNAL_PORT_NAME",
			Value: fmt.Sprintf("%s", DefaultInternalPortName),
		},
		corev1.EnvVar{
			Name:  "EXTERNAL_PORT_NAME",
			Value: fmt.Sprintf("%s", DefaultExternalPortName),
		},
		corev1.EnvVar{
			Name:  "INTERNAL_PORT",
			Value: fmt.Sprintf("%d", DefaultInternalPort),
		},
		corev1.EnvVar{
			Name:  "EXTERNAL_PORT",
			Value: fmt.Sprintf("%d", DefaultExternalPort),
		},
	)
	return envs
}
