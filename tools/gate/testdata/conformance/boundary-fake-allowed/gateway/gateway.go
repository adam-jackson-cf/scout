package gateway

type Gateway interface {
	Charge(cents int) error
}

func Bill(g Gateway, cents int) bool { return g.Charge(cents) == nil }
