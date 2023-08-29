package writer

import (
	"dior/option"
	"errors"
)

type WriteAble interface {
	Open() error
	Close() error
	Write(data string) error
}

type WriteCreator func(options *option.Options) (WriteAble, error)

var writerCreatorMap map[string]WriteCreator = make(map[string]WriteCreator)

func RegWriteCreator(dest string, creator WriteCreator) error {
	if oldCreator := writerCreatorMap[dest]; oldCreator != nil {
		return errors.New("conflict type : " + dest)
	}
	writerCreatorMap[dest] = creator
	return nil
}

func NewWriter(opts *option.Options) (WriteAble, error) {
	creator := writerCreatorMap[opts.Dest]
	if creator == nil {
		return nil, errors.New("unsupported type : " + opts.Dest)
	}
	return creator(opts)
}
