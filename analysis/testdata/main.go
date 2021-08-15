package main

func main() {
	One()
	Two()
	err := Three()
	_ = err
	anon := func() error {
		return nil
	}()
	err = anon
}

func One() {

}

func Two() {
	One()
}

func Three() error {
	return nil
}

type Foo struct {
	x string
}

func (strt *Foo) Bar() string {
	return strt.x
}
