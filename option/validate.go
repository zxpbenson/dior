package option

import (
	"errors"
	"os"
	"strings"
)

func (this *Options) Validate() error {
	if err := this.validateSrc(); err != nil {
		return err
	}

	if this.ChanSize < 0 {
		return errors.New("param [chan-size] : >= 0")
	}

	if err := this.validateDst(); err != nil {
		return err
	}
	return nil
}

func (this *Options) validateSrc() error {
	this.Src = strings.ToLower(this.Src)

	if this.Src != "nsq" && this.Src != "kafka" && this.Src != "press" {
		return errors.New("param [src] : nsq, kafka, press")
	}
	if this.Src == "kafka" {
		if err := this.validateSrcKafka(); err != nil {
			return err
		}
	}
	if this.Src == "nsq" {
		if err := this.validateSrcNSQ(); err != nil {
			return err
		}
	}
	if this.Src == "press" {
		if err := this.validateSrcPress(); err != nil {
			return err
		}
	}
	return nil
}

func (this *Options) validateSrcKafka() error {
	if len(this.SrcBootstrapServers) == 0 {
		return errors.New("param [src-bootstrap-servers] : required")
	}
	if this.SrcGroup == "" {
		return errors.New("param [src-group] : required")
	}
	if this.SrcTopic == "" {
		return errors.New("param [src-topic] : required")
	}
	return nil
}

func (this *Options) validateSrcNSQ() error {
	if len(this.SrcLookupdTCPAddresses) == 0 && len(this.SrcNSQDTCPAddresses) == 0 {
		return errors.New("param [src-lookupd-tcp-addresses] or [src-nsqd-tcp-addresses] : required")
	}
	if this.SrcChannel == "" {
		return errors.New("param [src-channel] : required")
	}
	if this.SrcTopic == "" {
		return errors.New("param [src-topic] : required")
	}
	return nil
}

func (this *Options) validateSrcPress() error {
	if this.SrcSpeed < 0 {
		return errors.New("param [src-speed] : >= 0")
	}
	if this.SrcFile == "" {
		return errors.New("param [src-file] : required")
	} else {
		fileInfo, err := os.Stat(this.SrcFile)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.New("SrcFile not exists : " + this.SrcFile)
			}
			return errors.Join(err, errors.New("SrcFile access error : "+this.DstFile))
		}
		if fileInfo.IsDir() {
			return errors.New("SrcFile It`s a directory : " + this.SrcFile)
		}
		if fileInfo.Size() < 2 {
			return errors.New("SrcFile It`s empty : " + this.SrcFile)
		}
	}
	return nil
}

func (this *Options) validateDst() error {
	this.Dst = strings.ToLower(this.Dst)
	if this.Dst != "nsq" && this.Dst != "kafka" && this.Dst != "nil" && this.Dst != "file" {
		return errors.New("param [dst] : nsq, kafka, nil, file")
	}

	if this.Dst == "kafka" {
		if err := this.validateDstKafka(); err != nil {
			return err
		}
	}
	if this.Dst == "nsq" {
		if err := this.validateDstNSQ(); err != nil {
			return err
		}
	}
	if this.Dst == "nil" {
	}
	if this.Dst == "file" {
		if err := this.validateDstFile(); err != nil {
			return err
		}
	}
	return nil
}

func (this *Options) validateDstKafka() error {
	if len(this.DstBootstrapServers) == 0 {
		return errors.New("param [dst-bootstrap-servers] : required")
	}
	if this.DstTopic == "" {
		return errors.New("param [dst-topic] : required")
	}
	return nil
}

func (this *Options) validateDstNSQ() error {
	if len(this.DstLookupdTCPAddresses) == 0 && len(this.DstNSQDTCPAddresses) == 0 {
		return errors.New("param [dst-lookupd-tcp-addresses] or [dst-nsqd-tcp-addresses] : required")
	}
	if this.DstTopic == "" {
		return errors.New("param [dst-topic] : required")
	}
	return nil
}

func (this *Options) validateDstFile() error {
	if this.DstFile == "" {
		return errors.New("param [dst-file] : required")
	} else {
		//fileInfo, err := os.Stat(this.DstFile)
		_, err := os.Stat(this.DstFile)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return errors.Join(err, errors.New("DstFile access error : "+this.DstFile))
		}
		return errors.New("DstFile exists : " + this.DstFile)
		//if fileInfo.IsDir() {
		//	return errors.New("DstFile It`s a directory : " + this.DstFile)
		//}
		//if fileInfo.Size() > 1 {
		//	return errors.New("DstFile`s not empty : " + this.DstFile)
		//}
	}
	if this.DstBufSizeByte < 0 {
		return errors.New("param [dst-buf-size-byte] : >= 0")
	}
	return nil
}
