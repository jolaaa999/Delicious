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
