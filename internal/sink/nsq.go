package sink

import (
	"context"
	"dior/component"
	"dior/internal/lg"
	"dior/option"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nsqio/go-nsq"
)

type nsqSink struct {
	*component.Asynchronizer

	producers            []*nsq.Producer
	topic                string
	nsqdTCPAddresses     []string
	lookupdHTTPAddresses []string
	nsqdLen              int
	nsqdIndex            atomic.Int64 // 使用atomic保证并发安全
}

func init() {
	component.RegCmpCreator("nsq-sink", newNSQSink)
}

func newNSQSink(name string, opts *option.Options) (component.Component, error) {
	return &nsqSink{
		Asynchronizer:        component.NewAsynchronizer(name),
		topic:                opts.DstTopic,
		nsqdTCPAddresses:     opts.DstNSQDTCPAddresses,
		lookupdHTTPAddresses: opts.DstLookupdHTTPAddresses,
	}, nil
}

// lookupdNSQDNodes 从 nsqlookupd 查询可用的 nsqd 节点
// 当前实现不支持nsqd节点在运行期间发生变动的情况
func (s *nsqSink) lookupdNSQDNodes() ([]string, error) {
	// 使用 map 根据 ip:port 去重
	nsqdAddrMap := make(map[string]struct{})

	for _, lookupdAddr := range s.lookupdHTTPAddresses {
		// nsqlookupd 的 HTTP 地址已经是 HTTP 格式，直接使用
		// 确保地址有 http:// 前缀
		httpAddr := lookupdAddr
		if !strings.HasPrefix(httpAddr, "http://") && !strings.HasPrefix(httpAddr, "https://") {
			httpAddr = "http://" + httpAddr
		}

		url := fmt.Sprintf("%s/nodes", httpAddr)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			lg.DftLgr.Warn("nsqSink.lookupdNSQDNodes failed to query lookupd %s: %v", lookupdAddr, err)
			continue
		}
		defer resp.Body.Close()

		var result struct {
			Producers []struct {
				RemoteAddress    string `json:"remote_address"`
				Hostname         string `json:"hostname"`
				BroadcastAddress string `json:"broadcast_address"`
				TCPPort          int    `json:"tcp_port"`
				HTTPPort         int    `json:"http_port"`
			} `json:"producers"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			lg.DftLgr.Warn("nsqSink.lookupdNSQDNodes failed to decode response from %s: %v", lookupdAddr, err)
			continue
		}

		for _, producer := range result.Producers {
			addr := producer.BroadcastAddress
			if addr == "" {
				addr = producer.Hostname
			}
			nsqdAddr := fmt.Sprintf("%s:%d", addr, producer.TCPPort)
			nsqdAddrMap[nsqdAddr] = struct{}{}
		}
	}

	// 将 map 的 key 转换为切片
	nsqdAddresses := make([]string, 0, len(nsqdAddrMap))
	for addr := range nsqdAddrMap {
		nsqdAddresses = append(nsqdAddresses, addr)
	}

	return nsqdAddresses, nil
}

func (s *nsqSink) Init(channel chan []byte) (err error) {
	s.Asynchronizer.Init(channel)

	cfg := nsq.NewConfig()

	// 根据 lookupdTCPAddresses 或 nsqdTCPAddresses 初始化 producers
	// 这两个参数是互斥的，只会传入其中一个
	if len(s.lookupdHTTPAddresses) > 0 {
		// 通过 lookupd 发现 nsqd 节点
		nsqdAddresses, err := s.lookupdNSQDNodes()
		if err != nil {
			return fmt.Errorf("nsqSink.Init failed to discover nsqd from lookupd: %v", err)
		}
		if len(nsqdAddresses) == 0 {
			return fmt.Errorf("nsqSink.Init: no nsqd nodes discovered from lookupd")
		}
		s.nsqdTCPAddresses = nsqdAddresses
		lg.DftLgr.Info("nsqSink.Init discovered %d nsqd nodes from lookupd: %v", len(nsqdAddresses), nsqdAddresses)
	} else if len(s.nsqdTCPAddresses) == 0 {
		return fmt.Errorf("nsqSink.Init: no nsqd or lookupd addresses provided")
	}

	s.nsqdLen = len(s.nsqdTCPAddresses)
	s.producers = make([]*nsq.Producer, s.nsqdLen)
	for idx, nsqdTCPAddress := range s.nsqdTCPAddresses {
		s.producers[idx], err = nsq.NewProducer(nsqdTCPAddress, cfg)
		if err != nil {
			return fmt.Errorf("nsqSink.Init failed to create producer for %s: %v", nsqdTCPAddress, err)
		}
	}
	s.Output = s.output
	lg.DftLgr.Info("nsqSink.Init connected to %d nsqd nodes: %v", s.nsqdLen, s.nsqdTCPAddresses)
	return nil
}

func (s *nsqSink) output(data []byte) error {
	// 轮询选择producer（原子操作保证并发安全）
	index := s.nsqdIndex.Add(1) - 1
	producer := s.producers[index%int64(s.nsqdLen)]

	// 发送消息
	err := producer.Publish(s.topic, data)
	if err == nil {
		lg.DftLgr.Debug("nsqSink.output publish ok, topic: %s", s.topic)
		return nil
	}
	return fmt.Errorf("nsqSink.output publish failed, topic: %s, error: %v", s.topic, err)
}

func (s *nsqSink) Start(ctx context.Context) {
	s.Asynchronizer.Start(ctx)
}

func (s *nsqSink) Stop() {
	for _, producer := range s.producers {
		producer.Stop()
	}
	s.Asynchronizer.Stop()
	lg.DftLgr.Info("nsqSink.Stop done.")
}
