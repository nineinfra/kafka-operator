package controller

import (
	"fmt"
	"strconv"
)

const (
	// DefaultNameSuffix is the default name suffix of the resources of the kafka
	DefaultNameSuffix = "-kafka"

	// DefaultClusterSign is the default cluster sign of the kafka
	DefaultClusterSign = "kafka"

	// DefaultStorageClass is the default storage class of the kafka
	DefaultStorageClass = "nineinfra-default"

	// DefaultReplicas is the default replicas
	DefaultReplicas = 3

	// DefaultDiskNum is the default disk num
	DefaultDiskNum = 1

	DefaultClusterDomainName = "clusterDomain"
	DefaultClusterDomain     = "cluster.local"

	//DefaultDataVolumeName = "data"

	DefaultLogVolumeName = "log"

	DefaultConfigNameSuffix      = "-config"
	DefaultHeadlessSvcNameSuffix = "-headless"

	DefaultKafkaHome           = "/opt/kafka"
	DefaultKafkaConfigFileName = "server.properties"
	DefaultLogConfigFileName   = "log4j.properties"
	DefaultDiskPathPrefix      = "disk"

	DefaultMaxBrokerID                = -1
	DefaultNetworkThreads             = 3
	DefaultIOThreads                  = 8
	DefaultSocketSendBufferSize       = 102400
	DefaultSocketReceiveBufferSize    = 102400
	DefaultSocketRequestsMaxSize      = 104857600
	DefaultPartitionsPerTopic         = 1
	DefaultRecoveryThreadsPerDir      = 1
	DefaultOffsetsFactor              = 1
	DefaultTransactionFactor          = 1
	DefaultTransactionISR             = 1
	DefaultLogFlushMessages           = 10000
	DefaultLogFlushInterval           = 1000
	DefaultLogRetentionHours          = 168
	DefaultLogSegmentSize             = 1073741824
	DefaultLogRetentionCheckInterval  = 300000
	DefaultZKConnectionTimeout        = 18000
	DefaultGroupInitialRebalanceDelay = 0

	DefaultInternalPortName = "internal"
	DefaultInternalPort     = 9092

	DefaultExternalPortName = "external"
	DefaultExternalPort     = 9093

	// DefaultTerminationGracePeriod is the default time given before the
	// container is stopped. This gives clients time to disconnect from a
	// specific node gracefully.
	DefaultTerminationGracePeriod = 30

	// DefaultKafkaVolumeSize is the default volume size for the
	// Kafka cache volume
	DefaultKafkaVolumeSize    = "50Gi"
	DefaultKafkaLogVolumeSize = "5Gi"

	// DefaultReadinessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the readiness probe
	DefaultReadinessProbeInitialDelaySeconds = 40

	// DefaultReadinessProbePeriodSeconds is the default probe period (in seconds)
	// for the readiness probe
	DefaultReadinessProbePeriodSeconds = 10

	// DefaultReadinessProbeFailureThreshold is the default probe failure threshold
	// for the readiness probe
	DefaultReadinessProbeFailureThreshold = 10

	// DefaultReadinessProbeSuccessThreshold is the default probe success threshold
	// for the readiness probe
	DefaultReadinessProbeSuccessThreshold = 1

	// DefaultReadinessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the readiness probe
	DefaultReadinessProbeTimeoutSeconds = 10

	// DefaultLivenessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the liveness probe
	DefaultLivenessProbeInitialDelaySeconds = 40

	// DefaultLivenessProbePeriodSeconds is the default probe period (in seconds)
	// for the liveness probe
	DefaultLivenessProbePeriodSeconds = 10

	// DefaultLivenessProbeFailureThreshold is the default probe failure threshold
	// for the liveness probe
	DefaultLivenessProbeFailureThreshold = 10

	// DefaultLivenessProbeSuccessThreshold is the default probe success threshold
	// for the readiness probe
	DefaultLivenessProbeSuccessThreshold = 1

	// DefaultLivenessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the liveness probe
	DefaultLivenessProbeTimeoutSeconds = 10

	//DefaultProbeTypeLiveness liveness type probe
	DefaultProbeTypeLiveness = "liveness"

	//DefaultProbeTypeReadiness readiness type probe
	DefaultProbeTypeReadiness = "readiness"
)

var (
	DefaultConfPath = fmt.Sprintf("%s/%s", DefaultKafkaHome, "conf")
	DefaultDataPath = fmt.Sprintf("%s/%s", DefaultKafkaHome, "data")
	DefaultLogPath  = fmt.Sprintf("%s/%s", DefaultKafkaHome, "logs")
)
var DefaultClusterConfKeyValue = map[string]string{
	"broker.id": strconv.Itoa(DefaultMaxBrokerID),
	//"log.dirs":                                 DefaultDataPath,
	//"listeners":                                "",
	//"advertised.listeners":                     "",
	"listener.security.protocol.map":           "PLAINTEXT:PLAINTEXT,SSL:SSL,SASL_PLAINTEXT:SASL_PLAINTEXT,SASL_SSL:SASL_SSL",
	"num.network.threads":                      strconv.Itoa(DefaultNetworkThreads),
	"num.io.threads":                           strconv.Itoa(DefaultIOThreads),
	"socket.send.buffer.bytes":                 strconv.Itoa(DefaultSocketSendBufferSize),
	"socket.receive.buffer.bytes":              strconv.Itoa(DefaultSocketReceiveBufferSize),
	"socket.request.max.bytes":                 strconv.Itoa(DefaultSocketRequestsMaxSize),
	"num.partitions":                           strconv.Itoa(DefaultPartitionsPerTopic),
	"num.recovery.threads.per.data.dir":        strconv.Itoa(DefaultRecoveryThreadsPerDir),
	"offsets.topic.replication.factor":         strconv.Itoa(DefaultOffsetsFactor),
	"transaction.state.log.replication.factor": strconv.Itoa(DefaultTransactionFactor),
	"transaction.state.log.min.isr":            strconv.Itoa(DefaultTransactionISR),
	"log.flush.interval.messages":              strconv.Itoa(DefaultLogFlushMessages),
	"log.flush.interval.ms":                    strconv.Itoa(DefaultLogFlushInterval),
	"log.retention.hours":                      strconv.Itoa(DefaultLogRetentionHours),
	"log.segment.bytes":                        strconv.Itoa(DefaultLogSegmentSize),
	"log.retention.check.interval.ms":          strconv.Itoa(DefaultLogRetentionCheckInterval),
	//"zookeeper.connect":                        "",
	"zookeeper.connection.timeout.ms":  strconv.Itoa(DefaultZKConnectionTimeout),
	"group.initial.rebalance.delay.ms": strconv.Itoa(DefaultGroupInitialRebalanceDelay),
}

var DefaultLogConfKeyValue = map[string]string{
	"log4j.rootLogger":                               "INFO, stdout, kafkaAppender",
	"log4j.appender.stdout":                          "org.apache.log4j.ConsoleAppender",
	"log4j.appender.stdout.layout":                   "org.apache.log4j.PatternLayout",
	"log4j.appender.stdout.layout.ConversionPattern": "[%d] %p %m (%c)%n",

	"log4j.appender.kafkaAppender":                          "org.apache.log4j.DailyRollingFileAppender",
	"log4j.appender.kafkaAppender.DatePattern":              "'.'yyyy-MM-dd-HH",
	"log4j.appender.kafkaAppender.File":                     "${kafka.logs.dir}/server.log",
	"log4j.appender.kafkaAppender.layout":                   "org.apache.log4j.PatternLayout",
	"log4j.appender.kafkaAppender.layout.ConversionPattern": "[%d] %p %m (%c)%n",
}
