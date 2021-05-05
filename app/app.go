package app

import "fmt"

var nami *Nami

func init() {

}

func NewNami() {
	if nami != nil {
		fmt.Println("created")
		return
	}

	fmt.Println("create")
	nami = &Nami{
		A: 2,
		B: 2,
	}
}

func NamiA() int {
	return nami.A
}

func NamiPlus() {
	nami.A++
}

type Nami struct {
	A int
	B int
}

func (n *Nami) Plus() {
	n.A++
}
