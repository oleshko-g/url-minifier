package reset_example

type myBool bool

// genStruct is an example of a resettable struct
//
// generate:reset
type genStruct struct {
	s  string
	sp *string
	i  int
	mb myBool
	sf []float64
	m  map[string]genStruct
	f  func()
	ch chan struct{}
	a  any
	st struct{}
	st2 struct{}
}
