package pipeline

import (
	"testing"
)

func TestSafeString(t *testing.T) {
	check := func(str string, expected string) {
		actual := safeString(str)
		if actual != expected {
			t.Errorf(`got %#v, wanted %#v`,
				actual, expected)
		}
	}

	for c := 1; c <= 31; c++ {

		switch c {
		case int('\t'), int('\n'):
			continue

		default:
			check((string)(c), "")
			check("A"+(string)(c)+"B", "AB")
		}
	}

	check("\t", "_")
	check("\n", "_")
	check("/", "-")
	check("\\", "-")
	check("*", "_")
	check(":", "_")
	check("?", "")
	check("<", "[")
	check(">", "]")
	check("\"", "'")

	check("A\tB", "A_B")
	check("A\nB", "A_B")
	check("A B", "A_B")
	check("A/B", "A-B")
	check("A\\B", "A-B")
	check("A*B", "A_B")
	check("A:B", "A_B")
	check("A?B", "AB")
	check("A<B", "A[B")
	check("A>B", "A]B")
	check("A\"B", "A'B")

	check(".", "_")
	check("..", "__")
	check("?.?", "_")
	check("?..?", "__")
	check("Русский текст", "Русский_текст")

}
