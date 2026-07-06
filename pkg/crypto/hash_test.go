package crypto

import "testing"

func TestSHA256Hex(t *testing.T) {
	got := SHA256Hex("abc")
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("SHA256Hex() = %q, want %q", got, want)
	}
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{name: "short", input: "abc", n: 8, want: "abc"},
		{name: "cut", input: "abcdef", n: 3, want: "abc"},
		{name: "zero", input: "abc", n: 0, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Prefix(tt.input, tt.n); got != tt.want {
				t.Fatalf("Prefix() = %q, want %q", got, tt.want)
			}
		})
	}
}
