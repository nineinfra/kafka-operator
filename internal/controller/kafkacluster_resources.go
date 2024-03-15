package controller

import (
	"fmt"
	kafkav1 "github.com/nineinfra/kafka-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

func volumeRequest(q resource.Quantity) corev1.ResourceList {
	m := make(corev1.ResourceList, 1)
	m[corev1.ResourceStorage] = q
	return m
}

func capacityPerVolume(capacity string) (*resource.Quantity, error) {
	totalQuantity, err := resource.ParseQuantity(capacity)
	if err != nil {
		return nil, err
	}
	return resource.NewQuantity(totalQuantity.Value(), totalQuantity.Format), nil
}

func getReplicas(cluster *kafkav1.KafkaCluster) int32 {
	if cluster.Spec.Resource.Replicas != 0 && cluster.Spec.Resource.Replicas%2 != 0 {
		return cluster.Spec.Resource.Replicas
	}
	return DefaultReplicas
}

func getDiskNum(cluster *kafkav1.KafkaCluster) int32 {
	if cluster.Spec.Resource.Disks != 0 {
		return cluster.Spec.Resource.Disks
	}

	return DefaultDiskNum
}

func getClusterConfigValue(cluster *kafkav1.KafkaCluster, key string, value string) string {
	if cluster.Spec.Conf != nil {
		if value, ok := cluster.Spec.Conf[key]; ok {
			return value
		}
	}
	return value
}

func constructClusterConfig(cluster *kafkav1.KafkaCluster) string {
	clusterConf := make(map[string]string)
	for k, v := range DefaultClusterConfKeyValue {
		clusterConf[k] = getClusterConfigValue(cluster, k, v)
	}
	if cluster.Spec.Conf != nil {
		for k, v := range cluster.Spec.Conf {
			if _, ok := DefaultClusterConfKeyValue[k]; !ok {
				clusterConf[k] = v
			}
		}
	}
	num := getDiskNum(cluster)
	volumeNames := make([]string, 0)
	for i := 0; i < int(num); i++ {
		volumeNames = append(volumeNames, fmt.Sprintf("%s/%s", DefaultDataPath, fmt.Sprintf("%s%d", DefaultDiskPathPrefix, i)))
	}
	clusterConf["log.dirs"] = strings.Join(volumeNames, ",")
	if _, ok := clusterConf["listeners"]; !ok {
		clusterConf["listeners"] = fmt.Sprintf("%s://0.0.0.0:%d,%s://0.0.0.0:%d",
			DefaultInternalPortName,
			DefaultInternalPort,
			DefaultExternalPortName,
			DefaultExternalPort)
		clusterConf["inter.broker.listener.name"] = DefaultInternalPortName
		clusterConf["listener.security.protocol.map"] = fmt.Sprintf("%s:PLAINTEXT,%s:PLAINTEXT",
			DefaultInternalPortName,
			DefaultExternalPortName)
	}

	return map2String(clusterConf)
}

func constructLogConfig() string {
	tmpConf := DefaultLogConfKeyValue
	return map2String(tmpConf)
}

func getImageConfig(cluster *kafkav1.KafkaCluster) kafkav1.ImageConfig {
	ic := kafkav1.ImageConfig{
		Repository:  cluster.Spec.Image.Repository,
		PullSecrets: cluster.Spec.Image.PullSecrets,
	}
	ic.Tag = cluster.Spec.Image.Tag
	if ic.Tag == "" {
		ic.Tag = cluster.Spec.Version
	}
	ic.PullPolicy = cluster.Spec.Image.PullPolicy
	if ic.PullPolicy == "" {
		ic.PullPolicy = string(corev1.PullIfNotPresent)
	}
	return ic
}

func (r *KafkaClusterReconciler) constructHeadlessService(cluster *kafkav1.KafkaCluster) (*corev1.Service, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(cluster, DefaultHeadlessSvcNameSuffix),
			Namespace: cluster.Namespace,
			Labels:    ClusterResourceLabels(cluster),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: DefaultInternalPortName,
					Port: DefaultInternalPort,
				},
				{
					Name: DefaultExternalPortName,
					Port: DefaultExternalPort,
				},
			},
			Selector:  ClusterResourceLabels(cluster),
			ClusterIP: corev1.ClusterIPNone,
		},
	}
	if err := ctrl.SetControllerReference(cluster, svc, r.Scheme); err != nil {
		return svc, err
	}
	return svc, nil
}

func (r *KafkaClusterReconciler) constructService(cluster *kafkav1.KafkaCluster) (*corev1.Service, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(cluster),
			Namespace: cluster.Namespace,
			Labels:    ClusterResourceLabels(cluster),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: DefaultExternalPortName,
					Port: DefaultExternalPort,
				},
				{
					Name: DefaultInternalPortName,
					Port: DefaultInternalPort,
				},
			},
			Selector: ClusterResourceLabels(cluster),
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	if err := ctrl.SetControllerReference(cluster, svc, r.Scheme); err != nil {
		return svc, err
	}
	return svc, nil
}

func (r *KafkaClusterReconciler) constructConfigMap(cluster *kafkav1.KafkaCluster) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(cluster, DefaultConfigNameSuffix),
			Namespace: cluster.Namespace,
			Labels:    ClusterResourceLabels(cluster),
		},
		Data: map[string]string{
			DefaultKafkaConfigFileName: constructClusterConfig(cluster),
			DefaultLogConfigFileName:   constructLogConfig(),
		},
	}
	if err := ctrl.SetControllerReference(cluster, cm, r.Scheme); err != nil {
		return cm, err
	}
	return cm, nil
}

func (r *KafkaClusterReconciler) getStorageRequests(cluster *kafkav1.KafkaCluster) (*resource.Quantity, error) {
	if cluster.Spec.Resource.ResourceRequirements.Requests != nil {
		if value, ok := cluster.Spec.Resource.ResourceRequirements.Requests["storage"]; ok {
			return &value, nil
		}
	}
	return capacityPerVolume(DefaultKafkaVolumeSize)
}

func (r *KafkaClusterReconciler) defaultKafkaPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          DefaultInternalPortName,
			ContainerPort: DefaultInternalPort,
		},
		{
			Name:          DefaultExternalPortName,
			ContainerPort: DefaultExternalPort,
		},
	}
}

func (r *KafkaClusterReconciler) constructVolumeMounts(cluster *kafkav1.KafkaCluster) []corev1.VolumeMount {
	num := int(getDiskNum(cluster))
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      ClusterResourceName(cluster, DefaultConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/conf/%s", DefaultKafkaHome, DefaultKafkaConfigFileName),
			SubPath:   DefaultKafkaConfigFileName,
		},
		{
			Name:      ClusterResourceName(cluster, DefaultConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/conf/%s", DefaultKafkaHome, DefaultLogConfigFileName),
			SubPath:   DefaultLogConfigFileName,
		},
		{
			Name:      DefaultLogVolumeName,
			MountPath: DefaultLogPath,
		},
	}
	for i := 0; i < num; i++ {
		volumeName := fmt.Sprintf("%s%d", DefaultDiskPathPrefix, i)
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: fmt.Sprintf("%s/%s", DefaultDataPath, volumeName),
		})
	}
	return volumeMounts
}

func (r *KafkaClusterReconciler) constructVolumes(cluster *kafkav1.KafkaCluster) []corev1.Volume {
	num := int(getDiskNum(cluster))
	volumes := []corev1.Volume{
		{
			Name: ClusterResourceName(cluster, DefaultConfigNameSuffix),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ClusterResourceName(cluster, DefaultConfigNameSuffix),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  DefaultKafkaConfigFileName,
							Path: DefaultKafkaConfigFileName,
						},
						{
							Key:  DefaultLogConfigFileName,
							Path: DefaultLogConfigFileName,
						},
					},
				},
			},
		},
	}

	volumes = append(volumes, corev1.Volume{
		Name: DefaultLogVolumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: DefaultLogVolumeName,
				ReadOnly:  false,
			},
		},
	})

	for i := 0; i < num; i++ {
		volumes = append(volumes, corev1.Volume{
			Name: fmt.Sprintf("%s%d", DefaultDiskPathPrefix, i),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("%s%d", DefaultDiskPathPrefix, i),
					ReadOnly:  false,
				},
			}},
		)
	}
	return volumes
}

func (r *KafkaClusterReconciler) constructPVCs(cluster *kafkav1.KafkaCluster) ([]corev1.PersistentVolumeClaim, error) {
	q, err := r.getStorageRequests(cluster)
	if err != nil {
		return nil, err
	}
	logq, err := capacityPerVolume(DefaultKafkaLogVolumeSize)
	if err != nil {
		return nil, err
	}
	sc := GetStorageClassName(cluster)

	num := int(getDiskNum(cluster))
	pvcs := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultLogVolumeName,
				Namespace: cluster.Namespace,
				Labels:    ClusterResourceLabels(cluster),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: &sc,
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: volumeRequest(*logq),
				},
			},
		},
	}
	for i := 0; i < num; i++ {
		pvcs = append(pvcs, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s%d", DefaultDiskPathPrefix, i),
				Namespace: cluster.Namespace,
				Labels:    ClusterResourceLabels(cluster),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: &sc,
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: volumeRequest(*q),
				},
			},
		})
	}
	return pvcs, nil
}

func (r *KafkaClusterReconciler) getProbeHandler(cluster *kafkav1.KafkaCluster, pType string) corev1.ProbeHandler {
	return corev1.ProbeHandler{
		TCPSocket: &corev1.TCPSocketAction{
			Port: intstr.FromInt32(DefaultInternalPort),
		},
	}
}

func (r *KafkaClusterReconciler) constructReadinessProbe(cluster *kafkav1.KafkaCluster) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        r.getProbeHandler(cluster, DefaultProbeTypeReadiness),
		InitialDelaySeconds: DefaultReadinessProbeInitialDelaySeconds,
		PeriodSeconds:       DefaultReadinessProbePeriodSeconds,
		TimeoutSeconds:      DefaultReadinessProbeTimeoutSeconds,
		FailureThreshold:    DefaultReadinessProbeFailureThreshold,
		SuccessThreshold:    DefaultReadinessProbeSuccessThreshold,
	}
}

func (r *KafkaClusterReconciler) constructLivenessProbe(cluster *kafkav1.KafkaCluster) *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        r.getProbeHandler(cluster, DefaultProbeTypeLiveness),
		InitialDelaySeconds: DefaultLivenessProbeInitialDelaySeconds,
		PeriodSeconds:       DefaultLivenessProbePeriodSeconds,
		TimeoutSeconds:      DefaultLivenessProbeTimeoutSeconds,
		FailureThreshold:    DefaultLivenessProbeFailureThreshold,
		SuccessThreshold:    DefaultLivenessProbeSuccessThreshold,
	}
}

func (r *KafkaClusterReconciler) constructKafkaPodSpec(cluster *kafkav1.KafkaCluster) corev1.PodSpec {
	tgp := int64(DefaultTerminationGracePeriod)
	ic := getImageConfig(cluster)
	var tmpPullSecrets []corev1.LocalObjectReference
	if ic.PullSecrets != "" {
		tmpPullSecrets = make([]corev1.LocalObjectReference, 0)
		tmpPullSecrets = append(tmpPullSecrets, corev1.LocalObjectReference{Name: ic.PullSecrets})
	}
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            cluster.Name,
				Image:           ic.Repository + ":" + ic.Tag,
				ImagePullPolicy: corev1.PullPolicy(ic.PullPolicy),
				Ports:           r.defaultKafkaPorts(),
				Env:             DefaultEnvs(),
				ReadinessProbe:  r.constructReadinessProbe(cluster),
				LivenessProbe:   r.constructLivenessProbe(cluster),
				VolumeMounts:    r.constructVolumeMounts(cluster),
			},
		},
		ImagePullSecrets:              tmpPullSecrets,
		RestartPolicy:                 corev1.RestartPolicyAlways,
		TerminationGracePeriodSeconds: &tgp,
		Volumes:                       r.constructVolumes(cluster),
		Affinity: &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: ClusterResourceLabels(cluster),
						},
					},
				},
			},
		},
	}
}
func (r *KafkaClusterReconciler) constructKafkaWorkload(cluster *kafkav1.KafkaCluster) (*appsv1.StatefulSet, error) {
	pvcs, err := r.constructPVCs(cluster)
	if err != nil {
		return nil, nil
	}
	stsDesired := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(cluster),
			Namespace: cluster.Namespace,
			Labels:    ClusterResourceLabels(cluster),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ClusterResourceLabels(cluster),
			},
			ServiceName: ClusterResourceName(cluster),
			Replicas:    int32Ptr(getReplicas(cluster)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ClusterResourceLabels(cluster),
				},
				Spec: r.constructKafkaPodSpec(cluster),
			},
			VolumeClaimTemplates: pvcs,
		},
	}

	if err := ctrl.SetControllerReference(cluster, stsDesired, r.Scheme); err != nil {
		return stsDesired, err
	}
	return stsDesired, nil
}
