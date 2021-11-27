package replay

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type mockFaultyWriter struct {
	*bytes.Buffer
	t           *testing.T
	called, mod int
}

func setupWriter(t *testing.T, mod int) *mockFaultyWriter {
	t.Helper()
	return &mockFaultyWriter{Buffer: new(bytes.Buffer), t: t, mod: mod}
}

func (m *mockFaultyWriter) Write(b []byte) (int, error) {
	defer func() { m.called += 1 }()
	if m.called%m.mod == 0 {
		return 0, fmt.Errorf("error")
	}
	return m.Buffer.Write(b)
}

func (m *mockFaultyWriter) reset() (io.Writer, error) {
	return m, nil
}

func (m *mockFaultyWriter) done(suffix string) {
	for !bytes.HasSuffix(m.Bytes(), []byte(suffix)) {
	}
	return
}

func (m *mockFaultyWriter) check(want string) {
	got := string(m.Bytes())
	if got != want {
		m.t.Fatalf("want: '%s', got: '%s'", want, got)
	}
}

func Test_Writer(t *testing.T) {
	cases := map[string]struct {
		mod, size int
		messages  []string
		want      string
	}{
		"base": {
			mod:      3,
			size:     10,
			messages: []string{"a", "b", "c"},
			want:     "abc",
		},
		"other": {
			mod:      2,
			size:     10,
			messages: []string{"aa", "bb", "\n"},
			want:     "aabb\n",
		},
	}
	for name := range cases {
		tc := cases[name]
		t.Run(name, func(t *testing.T) {
			tw := setupWriter(t, tc.mod)
			w, err := New(tw.reset, tc.size)
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range tc.messages {
				if _, err := w.Write([]byte(m)); err != nil {
					t.Fatal(err)
				}
				tw.done(m)
			}
			tw.check(tc.want)
		})
	}
}
