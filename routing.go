package rerr

func Tag(err *error, tag string) {

}

func TagByRouting(err *error, routes ...Route) {

}

type Route struct {
	Tag       string
	Predicate func(error) bool
}
