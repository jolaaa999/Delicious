package server

import (
	"net/http"

	"github.com/delicious/delicious/internal/app"
)

// Handler 供 Vercel Go Serverless 调用的 HTTP 入口。
func Handler(w http.ResponseWriter, r *http.Request) {
	app.Bootstrap().ServeHTTP(w, r)
}
