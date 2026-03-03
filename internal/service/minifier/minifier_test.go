package minifier

import "fmt"

func Example_encode() {
	const maxLen = 8
	out := encode([]byte("https://practicum.yandex.ru"), maxLen)
	fmt.Println(out)

	// Output:
	// a9tbDianbk0
}
