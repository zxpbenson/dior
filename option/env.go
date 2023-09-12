package option

import (
	"os"
	"strconv"
	"strings"
)

func (this *Options) loadEnv() (err error) {
	if logLevel := os.Getenv("log-level"); logLevel != "" {
		this.LogLevel = logLevel
	}

	if src := os.Getenv("src"); src != "" {
		this.Src = src
	}
	if srcTopic := os.Getenv("src-topic"); srcTopic != "" {
		this.SrcTopic = srcTopic
	}
	if srcLookupdTcpAddresses := os.Getenv("src-lookupd-tcp-addresses"); srcLookupdTcpAddresses != "" {
		this.SrcLookupdTCPAddresses = strings.Split(srcLookupdTcpAddresses, ",")
	}
	if srcNSQDTCPAddresses := os.Getenv("src-nsqd-tcp-addresses"); srcNSQDTCPAddresses != "" {
		this.SrcNSQDTCPAddresses = strings.Split(srcNSQDTCPAddresses, ",")
	}
	if srcChannel := os.Getenv("src-channel"); srcChannel != "" {
		this.SrcChannel = srcChannel
	}
	if srcBootstrapServers := os.Getenv("src-bootstrap-servers"); srcBootstrapServers != "" {
		this.SrcBootstrapServers = strings.Split(srcBootstrapServers, ",")
	}
	if srcGroup := os.Getenv("src-group"); srcGroup != "" {
		this.SrcGroup = srcGroup
	}
	if srcSpeed := os.Getenv("src-speed"); srcSpeed != "" {
		this.SrcSpeed, err = strconv.ParseInt(srcSpeed, 10, 64)
		if err != nil {
			return
		}
	}
	if srcFile := os.Getenv("src-file"); srcFile != "" {
		this.SrcFile = srcFile
	}

	if dst := os.Getenv("dst"); dst != "" {
		this.Dst = dst
	}
	if dstLookupdTCPAddresses := os.Getenv("dst-lookupd-tcp-address"); dstLookupdTCPAddresses != "" {
		this.DstLookupdTCPAddresses = strings.Split(dstLookupdTCPAddresses, ",")
	}
	if dstNSQDTCPAddresses := os.Getenv("dst-nsqd-tcp-address"); dstNSQDTCPAddresses != "" {
		this.DstNSQDTCPAddresses = strings.Split(dstNSQDTCPAddresses, ",")
	}
	if dstBootstrapServers := os.Getenv("dst-bootstrap-servers"); dstBootstrapServers != "" {
		this.DstBootstrapServers = strings.Split(dstBootstrapServers, ",")
	}
	if dstTopic := os.Getenv("dst-topic"); dstTopic != "" {
		this.DstTopic = dstTopic
	}
	if dstFile := os.Getenv("dst-file"); dstFile != "" {
		this.DstFile = dstFile
	}
	if dstBufSizeByte := os.Getenv("dst-buf-size-byte"); dstBufSizeByte != "" {
		this.DstBufSizeByte, err = strconv.Atoi(dstBufSizeByte)
		if err != nil {
			return
		}
	}
	return
}
