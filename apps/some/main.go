package main

import "fmt"

type I interface {
	F1()
	F2()
}

type ICreatorFunc func() I

var ICreatorFunMap map[string]ICreatorFunc = make(map[string]ICreatorFunc)

func init() {
	RegICreatorFunc("A", NewA)
	RegICreatorFunc("B", NewB)
}

type A struct {
}

func (this *A) F1() {
	fmt.Println("AAA f1")
}

func (this *A) F2() {
	fmt.Println("AAA f2")
}

type B struct {
	A
}

func (this *B) F2() {
	fmt.Println("BBB f2")
}

func RegICreatorFunc(name string, fun ICreatorFunc) {
	ICreatorFunMap[name] = fun
}

func NewI(name string) I {
	creator := ICreatorFunMap[name]
	return creator()
}

func NewA() I {
	return &A{}
}

func NewB() I {
	return &B{
		A{},
	}
}

func main() {
        fmt.Printf("start")
	var i1 I = &A{}
	var i2 I = &B{
		A{},
	}
	i1.F2()
	i2.F2()
}
