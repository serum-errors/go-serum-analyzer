package main

func main() {
	One()
	Two("shoop")
	err := Three()
	_ = err
	anon := func() error {
		return nil
	}()
	err = anon
	bark := (&Foo{"xoo"}).Bar("zinc")
	_ = bark
}

func One() {

}

func Two(zook string) {
	One()
}

func Three() error {
	return nil
}

type Foo struct {
	x string
}

func (strt *Foo) Bar(zonk string) string {
	return zonk + strt.x
}
