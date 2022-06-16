package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var uifiles embed.FS

type Ui struct {
	handler    http.ServeMux
	httpServer http.Server
}

func (u *Ui) init() {
	tfs, _ := fs.Sub(uifiles, "dist")
	u.handler.Handle("/", http.FileServer(http.FS(tfs)))
	u.handler.Handle("/api/servers", u.cross(u.apiServers))

	u.httpServer.Handler = &u.handler
}

func (u *Ui) cross(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if !crossRegexp.MatchString(origin) {
				w.WriteHeader(403)
				return
			}

			//w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Origin", origin)

			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json; charset=utf-8")

		fn(w, r)
	}
}

//读取状态
func (u *Ui) apiServers(w http.ResponseWriter, r *http.Request) {
	s := struct {
	}{}

	json.NewEncoder(w).Encode(&s)
}
