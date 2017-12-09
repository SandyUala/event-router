package prom

import (
	"encoding/json"
	"strconv"

	"github.com/astronomerio/event-router/kafka"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	log                     = logrus.WithField("package", "prom")
	statsMetadataCacheCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafka_metadata_cache_cnt",
		Help: "Kafka Metadata Cache Count",
	}, []string{"name", "client"})
	statsMessageCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafka_msg_cnt",
		Help: "Kafka Message Count",
	}, []string{"name", "client"})
	statsMessageMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafka_msg_max",
		Help: "Kafka Message Max",
	}, []string{"name", "client"})
	statsMessageSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafka_msg_size",
		Help: "Kafak Message Size",
	}, []string{"name", "client"})
	statsMessageSizeMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafka_message_size_max",
		Help: "Kafak Message Size Max",
	}, []string{"name", "client"})
	statsReplyQ = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafak_replyq",
		Help: "Kafka replyq",
	}, []string{"name", "client"})
	statsSimpleCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafka_simple_cnt",
		Help: "Kafka Simple Count",
	}, []string{"name", "client"})
	statsTS = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_kafka_ts",
		Help: "Kafka TS",
	}, []string{"name", "client"})

	brokerBufGrow = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_buf_grow",
		Help: "Kafka Broker Buf Grow",
	}, []string{"name", "nodeId", "client"})
	brokerIntLatencyAvg = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_int_latency_avg",
		Help: "Kafka Broker Int Latency Average",
	}, []string{"name", "nodeId", "client"})
	brokerIntLatencyCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_int_latency_cnt",
		Help: "Kafka Broker Int Latency Count",
	}, []string{"name", "nodeId", "client"})
	brokerIntLatencyMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_int_latency_max",
		Help: "Kafka Broker Int Latency Max",
	}, []string{"name", "nodeId", "client"})
	brokerIntLatencyMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_int_latency_min",
		Help: "Kafka Broker Int Latency Min",
	}, []string{"name", "nodeId", "client"})
	brokerIntLatencySum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_int_latency_sum",
		Help: "Kafka Broker Int Latency Sum",
	}, []string{"name", "nodeId", "client"})
	brokerOutBufCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_out_buf_cnt",
		Help: "Kafka Broker Out Buffer Count",
	}, []string{"name", "nodeId", "client"})
	brokerOutBufMsgCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_out_buf_msg_cnt",
		Help: "Kafka Broker Out Buffer Message Count",
	}, []string{"name", "nodeId", "client"})
	brokerReqTimeouts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_reg_timeouts",
		Help: "Kafka Broker Request Timeouts",
	}, []string{"name", "nodeId", "client"})
	brokerRttAvg = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rtt_avg",
		Help: "Kafka Broker RTT Average",
	}, []string{"name", "nodeId", "client"})
	brokerRttCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rtt_cnt",
		Help: "Kafka Broker RTT Count",
	}, []string{"name", "nodeId", "client"})
	brokerRttMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rtt_max",
		Help: "Kafka Broker RTT Max",
	}, []string{"name", "nodeId", "client"})
	brokerRttMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rtt_min",
		Help: "Kafka Broker RTT Min",
	}, []string{"name", "nodeId", "client"})
	brokerRttSum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rtt_sum",
		Help: "Kafka Broker RTT Sum",
	}, []string{"name", "nodeId", "client"})
	brokerRx = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rx",
		Help: "Kafka Broker RX",
	}, []string{"name", "nodeId", "client"})
	brokerRxBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rx_bytes",
		Help: "Kafka Broker RX Bytes",
	}, []string{"name", "nodeId", "client"})
	brokerRxCorridErrs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rx_corriderrs",
		Help: "Kafka Broker RX Corriderrs",
	}, []string{"name", "nodeId", "client"})
	brokerRxErrs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rx_errs",
		Help: "Kafka Broker RX Errors",
	}, []string{"name", "nodeId", "client"})
	brokerRxPartial = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_rx_partial",
		Help: "Kafka Broker RX_partial",
	}, []string{"name", "nodeId", "client"})
	brokerThrottleAvg = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_throttle_avg",
		Help: "Kafka Broker Throttle Average",
	}, []string{"name", "nodeId", "client"})
	brokerThrottleCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_throttle_cnt",
		Help: "Kafka Broker Throttle Count",
	}, []string{"name", "nodeId", "client"})
	brokerThrottleMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_throttle_max",
		Help: "Kafka Broker Throttle Max",
	}, []string{"name", "nodeId", "client"})
	brokerThrottleMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_throttle_min",
		Help: "Kafka Broker Throttle Min",
	}, []string{"name", "nodeId", "client"})
	brokerThrottleSum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_throttle_sum",
		Help: "Kafka Broker Throttle Sum",
	}, []string{"name", "nodeId", "client"})
	brokerTx = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_tx",
		Help: "Kafka Broker TX",
	}, []string{"name", "nodeId", "client"})
	brokerTxBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_tx_bytes",
		Help: "Kafka Broker TX Bytes",
	}, []string{"name", "nodeId", "client"})
	brokerTxErrs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_tx_errs",
		Help: "Kafka Broker TX Errors",
	}, []string{"name", "nodeId", "client"})
	brokerTxRetries = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_tx_retries",
		Help: "Kafka Broker TX Retries",
	}, []string{"name", "nodeId", "client"})
	brokerWaitRespCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_wait_resp_cnt",
		Help: "Kafka Broker Wait Response Count",
	}, []string{"name", "nodeId", "client"})
	brokerWaitRespMsgCnt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_wait_resp_msg_cnt",
		Help: "Kafka Broker Wait Response Message Count",
	}, []string{"name", "nodeId", "client"})
	brokerWakeups = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_wakeups",
		Help: "Kafka Broker Wakeups",
	}, []string{"name", "nodeId", "client"})
	brokerZBufGrow = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "event_router_broker_zbuf_grow",
		Help: "Kafka Broker ZBuffer Grow",
	}, []string{"name", "nodeId", "client"})
)

func init() {
	prometheus.MustRegister(
		statsMetadataCacheCount,
		statsMessageCount,
		statsMessageMax,
		statsMessageSize,
		statsMessageSizeMax,
		statsReplyQ,
		statsSimpleCnt,
		brokerBufGrow,
		brokerIntLatencyAvg,
		brokerIntLatencyCnt,
		brokerIntLatencyMax,
		brokerIntLatencyMin,
		brokerIntLatencySum,
		brokerOutBufCnt,
		brokerOutBufMsgCnt,
		brokerReqTimeouts,
		brokerRttAvg,
		brokerRttCnt,
		brokerRttMax,
		brokerRttMin,
		brokerRttSum,
		brokerRx,
		brokerRxBytes,
		brokerRxCorridErrs,
		brokerRxErrs,
		brokerRxPartial,
		brokerThrottleAvg,
		brokerThrottleCnt,
		brokerThrottleMax,
		brokerThrottleMin,
		brokerThrottleSum,
	)
}

func HandleKafkaStats(ks *confluent.Stats, client string) {
	var stats kafka.Stats
	err := json.Unmarshal([]byte(ks.String()), &stats)
	if err != nil {
		log.Errorf("json unmarshal error: %s", err)
		return
	}
	statsMetadataCacheCount.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.MetadataCacheCnt))
	statsMessageCount.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.MsgCnt))
	statsMessageMax.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.MsgMax))
	statsMessageSize.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.MsgSize))
	statsMessageSizeMax.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.MsgSizeMax))
	statsReplyQ.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.Replyq))
	statsSimpleCnt.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.SimpleCnt))
	statsTS.With(prometheus.Labels{"name": stats.Name, "client": client}).Set(float64(stats.Ts))
	for _, broker := range stats.Brokers {
		nodeId := strconv.Itoa(broker.Nodeid)
		brokerBufGrow.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.BufGrow))
		brokerIntLatencyAvg.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.IntLatency.Avg))
		brokerIntLatencyCnt.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.IntLatency.Cnt))
		brokerIntLatencyMax.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.IntLatency.Max))
		brokerIntLatencyMin.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.IntLatency.Min))
		brokerIntLatencySum.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.IntLatency.Sum))
		brokerOutBufCnt.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.OutbufCnt))
		brokerOutBufMsgCnt.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.OutbufMsgCnt))
		brokerReqTimeouts.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.ReqTimeouts))
		brokerRttAvg.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rtt.Avg))
		brokerRttCnt.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rtt.Cnt))
		brokerRttMax.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rtt.Max))
		brokerRttMin.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rtt.Min))
		brokerRttSum.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rtt.Sum))
		brokerRx.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rx))
		brokerRxBytes.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rxbytes))
		brokerRxCorridErrs.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rxcorriderrs))
		brokerRxErrs.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rxerrs))
		brokerRxPartial.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Rxpartial))
		brokerThrottleAvg.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Throttle.Avg))
		brokerThrottleCnt.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Throttle.Cnt))
		brokerThrottleMax.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Throttle.Max))
		brokerThrottleMin.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Throttle.Min))
		brokerThrottleSum.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Throttle.Sum))
		brokerTx.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Tx))
		brokerTxBytes.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Txbytes))
		brokerTxErrs.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Txerrs))
		brokerTxRetries.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Txretries))
		brokerWaitRespCnt.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.WaitrespCnt))
		brokerWaitRespMsgCnt.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.WaitrespMsgCnt))
		brokerWakeups.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.Wakeups))
		brokerZBufGrow.With(prometheus.Labels{"name": broker.Name, "nodeId": nodeId, "client": client}).Set(float64(broker.ZbufGrow))
	}
}
