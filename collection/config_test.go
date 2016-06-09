package collection

import (
	"strings"
	"testing"
)

func TestStringList(t *testing.T) {
	str := "th\\,is, is, a, test"
	got := StringList(str)
	expect := []string{
		"th,is",
		"is",
		"a",
		"test",
	}
	if len(got) != len(expect) {
		t.Error("Did not get expected number of values")
	}
	for i, s := range expect {
		if s != got[i] {
			t.Error("Expected: " + s + " Got: " + got[i])
		}
	}
}

func TestBasicConfig(t *testing.T) {
	reader := strings.NewReader("id:1234\nfoo:bar")
	configs := make(map[string]string)
	ParseConfig(reader, configs)
	tests := []struct {
		key    string
		expect string
	}{
		{
			key:    "id",
			expect: "1234",
		}, {
			key:    "foo",
			expect: "bar",
		},
	}
	for _, test := range tests {
		got := configs[test.key]
		if test.expect != got {
			t.Error("Expected: " + test.expect + " Got: " + got + " for key: " + test.key)
		}
	}
}

func TestTrimWs(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "\nthis is a test\n",
			expect: "this is a test",
		}, {
			input:  "   foo    ",
			expect: "foo",
		},
	}
	for _, test := range tests {
		got := trimWs(test.input)
		if test.expect != got {
			t.Error("Expected: " + test.expect + " Got: " + got)
		}
	}
}

func TestConfigReader(t *testing.T) {
	filesystemA()
	configs, e := LoadConfig("config.txt")
	if e != nil {
		t.Error(e)
	}
	tests := []struct {
		key    string
		expect string
	}{
		{
			key:    "id",
			expect: "1234",
		}, {
			key:    "foo",
			expect: "bar",
		},
	}
	for _, test := range tests {
		got := configs[test.key]
		if test.expect != got {
			t.Error("Expected: " + test.expect + " Got: " + got + " for key: " + test.key)
		}
	}
}
