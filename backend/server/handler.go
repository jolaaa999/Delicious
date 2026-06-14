package server

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/delicious/delicious/internal/app"
)

// Handler 供 Vercel Go Serverless 调用的 HTTP 入口。
func Handler(w http.ResponseWriter, r *http.Request) {
	eng, err := app.Engine()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	normalizeRequestPath(r)
	eng.ServeHTTP(w, r)
}

// normalizeRequestPath 修复 Vercel rewrite 后路径被截断为 /api 的问题。
func normalizeRequestPath(r *http.Request) {
	for _, key := range []string{
		"X-Vercel-Original-Url",
		"X-Original-Url",
		"X-Forwarded-Uri",
		"X-Invoke-Path",
	} {
		if v := r.Header.Get(key); v != "" {
			if u, parseErr := url.Parse(v); parseErr == nil && u.Path != "" && u.Path != r.URL.Path {
				r.URL.Path = u.Path
				r.URL.RawPath = u.RawPath
				return
			}
		}
	}

	if r.RequestURI != "" {
		if u, parseErr := url.ParseRequestURI(r.RequestURI); parseErr == nil && u.Path != "" {
			r.URL.Path = u.Path
		}
	}
}
