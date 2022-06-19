package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"github.com/timshannon/bolthold"
)

//go:embed dist/*
var uifiles embed.FS

type Ui struct {
	handler    http.ServeMux
	httpServer http.Server
	storeFile  string
	store      *bolthold.Store
}

func (u *Ui) init() {
	tfs, _ := fs.Sub(uifiles, "dist")
	u.handler.Handle("/", http.FileServer(http.FS(tfs)))
	u.handler.Handle("/api/servers", u.cross(u.apiServers))
	u.handler.Handle("/api/server/add", u.cross(u.apiServerAdd))
	u.handler.Handle("/api/server/edit", u.cross(u.apiServerEdit))
	u.handler.Handle("/api/server/del", u.cross(u.apiServerDel))
	u.handler.Handle("/api/viewLogs", u.cross(u.apiViewLogs))
	u.handler.Handle("/api/viewLog/add", u.cross(u.apiViewLogAdd))
	u.handler.Handle("/api/viewLog/edit", u.cross(u.apiViewLogEdit))
	u.handler.Handle("/api/viewLog/del", u.cross(u.apiViewLogDel))

	u.httpServer.Handler = &u.handler

	u.storeFile = "logViewer.db"

	var err error
	u.store, err = bolthold.Open(u.storeFile, 0644, nil)
	if err != nil {
		log.Fatalln("bolthold.Open", err)
	}
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
	t1 := []ServerConfig{}

	err := u.store.Find(&t1, new(bolthold.Query))
	if err != nil {
		log.Println("log find", err)
	}

	json.NewEncoder(w).Encode(&t1)
}

func (u *Ui) apiServerAdd(w http.ResponseWriter, r *http.Request) {
	var rs ServerConfig

	rs.Addr = r.FormValue("Addr")
	rs.User = r.FormValue("User")
	rs.Passwd = r.FormValue("Passwd")
	rs.PrivateKey = r.FormValue("PrivateKey")
	rs.UsePwd = r.FormValue("UsePwd") == "1"
	rs.Note = r.FormValue("Note")

	err := u.store.Insert(bolthold.NextSequence(), &rs)
	if err != nil {
		log.Println("apiServerConfigAdd", err)
	}

	w.Write([]byte("ok"))
}

func (u *Ui) apiServerEdit(w http.ResponseWriter, r *http.Request) {
	var rs ServerConfig

	id, _ := strconv.Atoi(r.FormValue("ID"))

	rs.ID = uint64(id)
	rs.Addr = r.FormValue("Addr")
	rs.User = r.FormValue("User")
	rs.Passwd = r.FormValue("Passwd")
	rs.PrivateKey = r.FormValue("PrivateKey")
	rs.UsePwd = r.FormValue("UsePwd") == "1"
	rs.Note = r.FormValue("Note")

	err := u.store.Update(rs.ID, rs)
	if err != nil {
		log.Println("apiServerConfigEdit", err)
	}

	w.Write([]byte("ok"))
}

func (u *Ui) apiServerDel(w http.ResponseWriter, r *http.Request) {
	var rs ServerConfig

	id, _ := strconv.Atoi(r.FormValue("ID"))
	rs.ID = uint64(id)

	err := u.store.Delete(rs.ID, rs)
	if err != nil {
		log.Println("apiServerConfigDel", err)
	}

	w.Write([]byte("ok"))
}

func (u *Ui) apiViewLogs(w http.ResponseWriter, r *http.Request) {
	t1 := []ViewLog{}

	err := u.store.Find(&t1, new(bolthold.Query))
	if err != nil {
		log.Println("log find", err)
	}

	json.NewEncoder(w).Encode(&t1)
}

func (u *Ui) apiViewLogAdd(w http.ResponseWriter, r *http.Request) {
	var rs ViewLog

	id, _ := strconv.Atoi(r.FormValue("ID"))
	rs.ID = uint64(id)
	rs.Note = r.FormValue("Note")
	rs.Files = r.FormValue("Files")
	if r.FormValue("Separator") != "" {
		rs.Separator = r.FormValue("Separator")[0]
	}
	rs.LineMatch = r.FormValue("LineMatch")
	rs.Match = r.Form["Match"]
	rs.Decoder = r.FormValue("Decoder")
	rs.BeginTime, _ = strconv.ParseInt(r.FormValue("BeginTime"), 10, 64)
	rs.StopTime, _ = strconv.ParseInt(r.FormValue("StopTime"), 10, 64)

	err := u.store.Insert(bolthold.NextSequence(), &rs)
	if err != nil {
		log.Println("apiServerConfigAdd", err)
	}

	w.Write([]byte("ok"))
}

func (u *Ui) apiViewLogEdit(w http.ResponseWriter, r *http.Request) {
	var rs ViewLog

	id, _ := strconv.Atoi(r.FormValue("ID"))

	rs.ID = uint64(id)
	rs.Note = r.FormValue("Note")
	rs.Files = r.FormValue("Files")
	if r.FormValue("Separator") != "" {
		rs.Separator = r.FormValue("Separator")[0]
	}
	rs.LineMatch = r.FormValue("LineMatch")
	rs.Match = r.Form["Match"]
	rs.Decoder = r.FormValue("Decoder")
	rs.BeginTime, _ = strconv.ParseInt(r.FormValue("BeginTime"), 10, 64)
	rs.StopTime, _ = strconv.ParseInt(r.FormValue("StopTime"), 10, 64)

	err := u.store.Update(rs.ID, rs)
	if err != nil {
		log.Println("apiViewLogEdit", err)
	}

	w.Write([]byte("ok"))
}

func (u *Ui) apiViewLogDel(w http.ResponseWriter, r *http.Request) {
	var rs ViewLog

	id, _ := strconv.Atoi(r.FormValue("ID"))
	rs.ID = uint64(id)

	err := u.store.Delete(rs.ID, rs)
	if err != nil {
		log.Println("apiViewLogDel", err)
	}

	w.Write([]byte("ok"))
}
