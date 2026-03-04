package minifier

import (
	"fmt"
	"testing"
)

const maxLen = 8

func Example_encode() {
	out := encode([]byte("https://practicum.yandex.ru"), maxLen)
	fmt.Println(out)

	// Output:
	// a9tbDianbk0
}

func Benchmark_encode(b *testing.B) {
	b.Run("short URL len 12", func(b *testing.B) {
		for b.Loop() {
			in := []byte("http://x.com")
			out := encode(in, maxLen)
			_ = out
		}

	})

	b.Run("median URL len 58", func(b *testing.B) {
		for b.Loop() {
			// len(in)
			in := []byte("https://practicum.yandex.ru/catalog/programming/pro/paid/")
			out := encode(in, maxLen)
			_ = out
		}
	})

	b.Run("extremly long URL len 2005", func(b *testing.B) {

		for b.Loop() {
			// len(in) is close to 2005 characters
			in := []byte("https://example.com/search?q=wireless+noise-cancelling+headphones&filter%5Btag1%5D=xxxxxxxxxxxxxxxxxxxx&filter%5Btag2%5D=xxxxxxxxxxxxxxxxxxxx&filter%5Btag3%5D=xxxxxxxxxxxxxxxxxxxx&filter%5Btag4%5D=xxxxxxxxxxxxxxxxxxxx&state=%7B%22sort%22%3A%22relevance%22%2C%22page%22%3A1%2C%22filters%22%3A%5B%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB01%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB02%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB03%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB04%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB05%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB06%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB07%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB08%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB09%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB10%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB11%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB12%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB13%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB14%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB15%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB16%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB17%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB18%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB19%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB20%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB21%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB22%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB23%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB24%22%7D%2C%7B%22k%22%3A%22brand%22%2C%22v%22%3A%22BBBBBBBBBB25%22%7D%5D%7D&utm_source=newsletter&utm_medium=email&utm_campaign=spring_sale_2026&session_id=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&sig=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
			out := encode(in, maxLen)
			_ = out
		}
	})

}
