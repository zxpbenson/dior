package component

import (
	"dior/option"
	"errors"
	"strings"
)

type Component interface {
	Controllable
	Init() (err error)
	Start()
	Stop()
	SetChannel(chan []byte)
	WaitEmpty()
}

type OutputFunc func(data []byte)

type ComponentCreator func(options *option.Options) (Component, error)

var cmpCreatorMap map[string]ComponentCreator = make(map[string]ComponentCreator)

func RegCmpCreator(name string, creator ComponentCreator) error {
	name = strings.ToLower(name)
	if oldCreator := cmpCreatorMap[name]; oldCreator != nil {
		return errors.New("conflict type : " + name)
	}
	cmpCreatorMap[name] = creator
	return nil
}

func NewComponent(name string, opts *option.Options) (Component, error) {
	name = strings.ToLower(name)
	creator := cmpCreatorMap[name]
	if creator == nil {
		return nil, errors.New("unsupported type : " + name)
	}
	return creator(opts)
}
