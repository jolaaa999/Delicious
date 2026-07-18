package handler

import (
	"net/http"

	"github.com/delicious/delicious/server"
)

// Handler 是 Vercel Go Serverless 的入口。
func Handler(w http.ResponseWriter, r *http.Request) {
	server.Handler(w, r)
}
