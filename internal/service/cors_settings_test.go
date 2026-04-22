package service

import (
	"reflect"
	"testing"
)

func TestParseCORSAllowedOrigins(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty", "", nil},
		{"whitespace only", "   \t\n ", nil},
		{"single", "https://kite.example", []string{"https://kite.example"}},
		{"trailing slash trimmed", "https://kite.example/", []string{"https://kite.example"}},
		{
			"comma separated",
			"https://a.example, https://b.example ,https://c.example",
			[]string{"https://a.example", "https://b.example", "https://c.example"},
		},
		{
			"newline separated",
			"https://a.example\nhttps://b.example\r\nhttps://c.example",
			[]string{"https://a.example", "https://b.example", "https://c.example"},
		},
		{
			"mixed separators and empty tokens",
			" ,https://a.example,\nhttps://b.example,,, https://c.example/ \n",
			[]string{"https://a.example", "https://b.example", "https://c.example"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseCORSAllowedOrigins(tc.raw)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParseCORSAllowedOrigins(%q) = %#v, want %#v", tc.raw, got, tc.want)
			}
		})
	}
}
