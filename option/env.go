package option

import (
	"os"
	"strconv"
	"strings"
)

func (o *Options) loadEnv() (err error) {
	if logLevel := os.Getenv("log-level"); logLevel != "" {
		o.LogLevel = logLevel
	}

	if src := os.Getenv("src"); src != "" {
		o.Src = src
	}
	if srcTopic := os.Getenv("src-topic"); srcTopic != "" {
		o.SrcTopic = srcTopic
	}
	if srcLookupdHTTPAddresses := os.Getenv("src-lookupd-http-addresses"); srcLookupdHTTPAddresses != "" {
		o.SrcLookupdHTTPAddresses = strings.Split(srcLookupdHTTPAddresses, ",")
	}
	if srcNSQDTCPAddresses := os.Getenv("src-nsqd-tcp-addresses"); srcNSQDTCPAddresses != "" {
		o.SrcNSQDTCPAddresses = strings.Split(srcNSQDTCPAddresses, ",")
	}
	if srcChannel := os.Getenv("src-channel"); srcChannel != "" {
		o.SrcChannel = srcChannel
	}
	if srcBootstrapServers := os.Getenv("src-bootstrap-servers"); srcBootstrapServers != "" {
		o.SrcBootstrapServers = strings.Split(srcBootstrapServers, ",")
	}
	if srcGroup := os.Getenv("src-group"); srcGroup != "" {
		o.SrcGroup = srcGroup
	}
	if srcSpeed := os.Getenv("src-speed"); srcSpeed != "" {
		o.SrcSpeed, err = strconv.ParseInt(srcSpeed, 10, 64)
		if err != nil {
			return
		}
	}
	if srcFile := os.Getenv("src-file"); srcFile != "" {
		o.SrcFile = srcFile
	}

	if dst := os.Getenv("dst"); dst != "" {
		o.Dst = dst
	}
	if dstLookupdHTTPAddresses := os.Getenv("dst-lookupd-http-address"); dstLookupdHTTPAddresses != "" {
		o.DstLookupdHTTPAddresses = strings.Split(dstLookupdHTTPAddresses, ",")
	}
	if dstNSQDTCPAddresses := os.Getenv("dst-nsqd-tcp-address"); dstNSQDTCPAddresses != "" {
		o.DstNSQDTCPAddresses = strings.Split(dstNSQDTCPAddresses, ",")
	}
	if dstBootstrapServers := os.Getenv("dst-bootstrap-servers"); dstBootstrapServers != "" {
		o.DstBootstrapServers = strings.Split(dstBootstrapServers, ",")
	}
	if dstTopic := os.Getenv("dst-topic"); dstTopic != "" {
		o.DstTopic = dstTopic
	}
	if dstFile := os.Getenv("dst-file"); dstFile != "" {
		o.DstFile = dstFile
	}
	if dstBufSizeByte := os.Getenv("dst-buf-size-byte"); dstBufSizeByte != "" {
		o.DstBufSizeByte, err = strconv.Atoi(dstBufSizeByte)
		if err != nil {
			return
		}
	}
	return
}
