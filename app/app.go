package app

type Nami struct {
	A int
}

func NewNami() {
	return &Nami{
		A;5
	}
}

func (n *Nami) Plus() {
	n.A++
}