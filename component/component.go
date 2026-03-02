package component

import (
	"context"
	"dior/option"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Component interface {
	Controllable
	Init(channel chan []byte) (err error)
	Start(ctx context.Context)
	Stop()
}

type OutputFunc func(data []byte)

type ComponentCreator func(name string, options *option.Options) (Component, error)

var cmpCreatorMap map[string]ComponentCreator = make(map[string]ComponentCreator)

func RegCmpCreator(name string, creator ComponentCreator) {
	name = strings.ToLower(name)
	if oldCreator := cmpCreatorMap[name]; oldCreator != nil {
		fmt.Printf("conflict type : " + name)
		os.Exit(1) // 这种错误只能开发阶段处理，有错误直接退出
	}
	cmpCreatorMap[name] = creator
}

func NewComponent(name string, opts *option.Options) (Component, error) {
	name = strings.ToLower(name)
	creator := cmpCreatorMap[name]
	if creator == nil {
		return nil, errors.New("unsupported type : " + name)
	}
	return creator(name, opts)
}
