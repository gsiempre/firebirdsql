package firebirdsql

import "testing"

func TestParseDSNPageSize(t *testing.T) {
	for dsns, want := range map[string]string{
		"user:pass@localhost/test.fdb":                 "4096",
		"user:pass@localhost/test.fdb?page_size=16384": "16384",
	} {
		dsn, err := parseDSN(dsns)
		if err != nil {
			t.Fatal(err)
		}
		if got := dsn.options["page_size"]; got != want {
			t.Errorf("%s: page_size = %q, want %q", dsns, got, want)
		}
	}
}

func TestParseDSNInvalidPageSize(t *testing.T) {
	for _, v := range []string{"4097", "0", "-4096", "1024", "2048", "65536", "abc", ""} {
		if _, err := parseDSN("user:pass@localhost/test.fdb?page_size=" + v); err == nil {
			t.Errorf("page_size=%q: expected error", v)
		}
	}
}
