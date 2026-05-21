package mykafka

import "time"

const (
	defaultPartition = 0

	producerWriteTimeout = 10 * time.Second
	producerReadTimeout  = 10 * time.Second
	producerBatchSize    = 10
	producerBatchBytes   = 1048576
	producerBatchTimeout = 100 * time.Millisecond

	readerMinBytes = 1
	readerMaxBytes = 10e6
	readerMaxWait  = 500 * time.Millisecond

	DefaultConsumerWorkerCount = 10
	DefaultConsumerQueueSize   = 1000

	consumerGroupStabilizeWait = 3 * time.Second
	fetcherStartInterval       = 500 * time.Millisecond
	partitionFetchRetryWait    = 2 * time.Second
)
