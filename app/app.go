package app

var nami *Nami

func init() {
	nami = &Nami{
		A: 2,
		B: 2,
	}
}

func NewNami() *Nami {
	return &Nami{
		A: 5,
		B: 5,
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
