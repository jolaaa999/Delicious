package service

import (
	"context"
	"net/http"
	"testing"
)

func TestExpandSearchKeywords_PorkRibs(t *testing.T) {
	keywords := expandSearchKeywords(context.Background(), http.DefaultClient, "排骨")
	if len(keywords) < 2 {
		t.Fatalf("expected at least 2 keywords, got %v", keywords)
	}
	found := false
	for _, kw := range keywords {
		if kw == "pork ribs" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected pork ribs in %v", keywords)
	}
}

func TestExpandSearchKeywords_EnglishPassthrough(t *testing.T) {
	keywords := expandSearchKeywords(context.Background(), http.DefaultClient, "pasta")
	if len(keywords) != 1 || keywords[0] != "pasta" {
		t.Fatalf("unexpected keywords: %v", keywords)
	}
}

func TestIsSuspiciousTranslation_SpamAndPlaceholder(t *testing.T) {
	cases := []struct {
		src, dst string
		want     bool
	}{
		{"cup", "千锋教育现有员工300多名，拥有全国最多", true},
		{"clove", "待定", true},
		{"onion", "洋葱", false},
		{"cup", "杯", false},
		{"tbsp", "MYMEMORY WARNING: YOU USED ALL AVAILABLE FREE TRANSLATIONS", true},
	}
	for _, c := range cases {
		if got := isSuspiciousTranslation(c.src, c.dst); got != c.want {
			t.Fatalf("isSuspiciousTranslation(%q, %q)=%v want %v", c.src, c.dst, got, c.want)
		}
	}
}

func TestLocalizeUnit_EnToZh(t *testing.T) {
	cases := map[string]string{
		"cup":    "杯",
		"cloves": "瓣",
		"tbsp":   "汤匙",
		"份":      "份",
		"":       "适量",
		"待定":     "适量",
		"千锋教育现有员工300多名拥有全国最多最权威": "适量",
	}
	for in, want := range cases {
		if got := localizeUnit(in, "en|zh-CN"); got != want {
			t.Fatalf("localizeUnit(%q)=%q want %q", in, got, want)
		}
	}
}
