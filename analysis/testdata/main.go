package main

func main() {
	One()
	Two()
	err := Three()
	_ = err
}

func One() {

}

func Two() {
	One()
}

func Three() error {
	return nil
}
