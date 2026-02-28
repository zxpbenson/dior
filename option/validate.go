package option

import (
	"errors"
	"os"
	"strings"
)

func (o *Options) Validate() error {
	if err := o.validateSrc(); err != nil {
		return err
	}

	if o.ChanSize < 0 {
		return errors.New("param [chan-size] : >= 0")
	}

	if err := o.validateDst(); err != nil {
		return err
	}
	return nil
}

func (o *Options) validateSrc() error {
	o.Src = strings.ToLower(o.Src)

	if o.Src != "nsq" && o.Src != "kafka" && o.Src != "press" {
		return errors.New("param [src] : nsq, kafka, press")
	}
	if o.Src == "kafka" {
		if err := o.validateSrcKafka(); err != nil {
			return err
		}
	}
	if o.Src == "nsq" {
		if err := o.validateSrcNSQ(); err != nil {
			return err
		}
	}
	if o.Src == "press" {
		if err := o.validateSrcPress(); err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) validateSrcKafka() error {
	if len(o.SrcBootstrapServers) == 0 {
		return errors.New("param [src-bootstrap-servers] : required")
	}
	if o.SrcGroup == "" {
		return errors.New("param [src-group] : required")
	}
	if o.SrcTopic == "" {
		return errors.New("param [src-topic] : required")
	}
	return nil
}

func (o *Options) validateSrcNSQ() error {
	if len(o.SrcLookupdTCPAddresses) == 0 && len(o.SrcNSQDTCPAddresses) == 0 {
		return errors.New("param [src-lookupd-tcp-addresses] or [src-nsqd-tcp-addresses] : required")
	}
	if o.SrcChannel == "" {
		return errors.New("param [src-channel] : required")
	}
	if o.SrcTopic == "" {
		return errors.New("param [src-topic] : required")
	}
	return nil
}

func (o *Options) validateSrcPress() error {
	if o.SrcSpeed < 0 {
		return errors.New("param [src-speed] : >= 0")
	}
	if o.SrcFile == "" {
		return errors.New("param [src-file] : required")
	} else {
		fileInfo, err := os.Stat(o.SrcFile)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.New("SrcFile not exists : " + o.SrcFile)
			}
			return errors.Join(err, errors.New("SrcFile access error : "+o.SrcFile))
		}
		if fileInfo.IsDir() {
			return errors.New("SrcFile It`s a directory : " + o.SrcFile)
		}
		if fileInfo.Size() < 2 {
			return errors.New("SrcFile It`s empty : " + o.SrcFile)
		}
	}
	return nil
}

func (o *Options) validateDst() error {
	o.Dst = strings.ToLower(o.Dst)
	if o.Dst != "nsq" && o.Dst != "kafka" && o.Dst != "nil" && o.Dst != "file" {
		return errors.New("param [dst] : nsq, kafka, nil, file")
	}

	if o.Dst == "kafka" {
		if err := o.validateDstKafka(); err != nil {
			return err
		}
	}
	if o.Dst == "nsq" {
		if err := o.validateDstNSQ(); err != nil {
			return err
		}
	}
	if o.Dst == "nil" {
	}
	if o.Dst == "file" {
		if err := o.validateDstFile(); err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) validateDstKafka() error {
	if len(o.DstBootstrapServers) == 0 {
		return errors.New("param [dst-bootstrap-servers] : required")
	}
	if o.DstTopic == "" {
		return errors.New("param [dst-topic] : required")
	}
	return nil
}

func (o *Options) validateDstNSQ() error {
	if len(o.DstLookupdTCPAddresses) == 0 && len(o.DstNSQDTCPAddresses) == 0 {
		return errors.New("param [dst-lookupd-tcp-addresses] or [dst-nsqd-tcp-addresses] : required")
	}
	if o.DstTopic == "" {
		return errors.New("param [dst-topic] : required")
	}
	return nil
}

func (o *Options) validateDstFile() error {
	if o.DstFile == "" {
		return errors.New("param [dst-file] : required")
	} else {
		_, err := os.Stat(o.DstFile)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return errors.Join(err, errors.New("DstFile access error : "+o.DstFile))
		}
		return errors.New("DstFile exists : " + o.DstFile)
	}
	if o.DstBufSizeByte < 0 {
		return errors.New("param [dst-buf-size-byte] : >= 0")
	}
	return nil
}
