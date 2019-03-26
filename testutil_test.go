package failure_test

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func shouldContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("%q does not contain %q", s, sub)
	}
}

func shouldMatch(t *testing.T, s, re string) {
	t.Helper()
	r := regexp.MustCompile(re)
	if !r.MatchString(s) {
		t.Errorf("%q does not match %q", s, re)
	}
}

func shouldEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%T(%#v) does not equal to %T(%#v)", a, a, b, b)
	}
}

func shouldDiffer(t *testing.T, a, b interface{}) {
	t.Helper()
	if reflect.DeepEqual(a, b) {
		t.Errorf("%T(%#v) does not differ from %T(%#v)", a, a, b, b)
	}
}
