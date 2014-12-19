package parse

import "testing"

func TestFirstToUpper(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"abcd", "Abcd"},
		{"fieldName", "FieldName"},
		{"OtherField", "OtherField"},
		{"Test", "Test"},
	}

	for _, c := range cases {
		actual := firstToUpper(c.input)
		if actual != c.expected {
			t.Errorf("unexpected output - got [%s], expected [%s]\n", actual, c.expected)
		}
	}
}

func BenchmarkFirstToUpper(b *testing.B) {
	for i := 0; i < b.N; i++ {
		firstToUpper("aLongFieldName")
	}
}

func TestFirstToLower(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"abcd", "abcd"},
		{"fieldName", "fieldName"},
		{"OtherField", "otherField"},
		{"Test", "test"},
	}

	for _, c := range cases {
		actual := firstToLower(c.input)
		if actual != c.expected {
			t.Errorf("unexpected output - got [%s], expected [%s]\n", actual, c.expected)
		}
	}
}

func BenchmarkFirstToLower(b *testing.B) {
	for i := 0; i < b.N; i++ {
		firstToLower("ALongFieldName")
	}
}
