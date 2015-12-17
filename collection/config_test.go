package collection

import (
	"fmt"
	"testing"
)

func TestStringList(t *testing.T) {
	s := "th\\,is, is, a, test"
	fmt.Println(StringList(s))
}
