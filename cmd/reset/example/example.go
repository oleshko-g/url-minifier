package example

type myBool bool

// genStruct is an example of a resettable struct
//
// generate:reset
type GenStruct struct {
	s  string
	sp *string
	i  int
	mb myBool
	sf []float64
	m  map[string]GenStruct
	f  func()
	ch chan struct{}
	a  any
	st struct{}
	st2 struct{}
}
