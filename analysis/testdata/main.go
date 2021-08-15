package main

import "fmt"

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

func Fun(flip bool) error {
	err := fmt.Errorf("asdf")
	if flip {
		err = fmt.Errorf("qwer")
	}
	return err
}

func ScaryShadowing(flip bool) error {
	err := fmt.Errorf("asdf")
	if flip {
		err := fmt.Errorf("qwer") // different scope!  cannot actually be returned.
		_ = err
	}
	return err
}

func Closures(flip bool) error {
	err := fmt.Errorf("asdf")
	if flip {
		func() {
			err = fmt.Errorf("qwer") // same value, just captured in a closure.
		}()
	}
	return err

}

func Trickier(flip bool) error {
	var foo struct {
		err error
	}
	foo.err = fmt.Errorf("asdf")
	if flip {
		foo.err = fmt.Errorf("qwer")
	}
	return foo.err
}

type Structural struct {
	Value interface{}
	Err   error
}

func Trickiest(flip bool) Structural {
	// I think we'll... not detect this by default.
	// But maybe you can put a comment on the struct field,
	// and we can trace assignments to that the same way as we trace starting from returns.
	return Structural{}
}
