package service

import (
	"fmt"
	"testing"
)

func TestIsProviderQuotaError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{fmt.Errorf("spoonacular: HTTP 402: daily points limit of 50 has been reached"), true},
		{fmt.Errorf("spoonacular: HTTP 429: rate limit"), true},
		{fmt.Errorf("spoonacular: HTTP 500: boom"), false},
		{fmt.Errorf("network timeout"), false},
	}
	for _, c := range cases {
		if got := isProviderQuotaError(c.err); got != c.want {
			t.Fatalf("isProviderQuotaError(%v)=%v want %v", c.err, got, c.want)
		}
	}
}
