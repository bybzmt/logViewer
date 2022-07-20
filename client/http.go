package client

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"logViewer/find"

	"github.com/gorilla/websocket"
	"github.com/timshannon/bolthold"
)

//go:embed dist/*
var uifiles embed.FS

type Ui struct {
	handler    http.ServeMux
	HttpServer http.Server
	storeFile  string
	store      *bolthold.Store
}

func (u *Ui) Init() {
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
	u.handler.Handle("/api/logs", u.cross(u.apiLogs))
	u.handler.Handle("/api/glob", u.cross(u.apiGlob))

	u.HttpServer.Handler = &u.handler

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

			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "content-type")
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
	rs.UsePwd = r.FormValue("UsePwd") == "true"
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
	rs.UsePwd = r.FormValue("UsePwd") == "true"
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

	err := u.store.Find(&t1, nil)
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
	rs.TimeRegex = r.FormValue("TimeRegex")
	rs.TimeLayout = r.FormValue("TimeLayout")
	rs.Contains = r.FormValue("Contains")
	rs.Regex = r.FormValue("Regex")

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
	rs.TimeRegex = r.FormValue("TimeRegex")
	rs.TimeLayout = r.FormValue("TimeLayout")
	rs.Contains = r.FormValue("Contains")
	rs.Regex = r.FormValue("Regex")

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

func CheckOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOrigin,
}

func (u *Ui) apiLogs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	rs, err := findDial(u.store, 1)
	if err != nil {
		log.Println("find", err)
		return
	}

	defer func() {
		if e := rs.Close(); e != nil {
			log.Println(e)
		}
	}()

	for {
		now := time.Now().Add(time.Second * 5)
		conn.SetReadDeadline(now)
		conn.SetWriteDeadline(now)

		mtype, buf, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage", err)
			return
		}

		var msg struct {
			Op   string
			Data json.RawMessage
		}

		err = json.Unmarshal(buf, &msg)
		if err != nil {
			log.Println("josn decode", err)
			return
		}

		if mtype == websocket.PingMessage {
			log.Println("ping")
		}

		if mtype == websocket.TextMessage {
			switch msg.Op {
			case "glob":
				var pattern string

				err = json.Unmarshal(msg.Data, &pattern)
				if err != nil {
					log.Println(err)
					return
				}

				files, err := rs.Glob(pattern)

				if err != nil {
					log.Println("glob", err)
					return
				}

				out := struct {
					Op    string
					Files []string
				}{Op: "glob", Files: files}

				if err := conn.WriteJSON(&out); err != nil {
					log.Println(err)
				}

			case "grep":
				var m find.MatchParam

				err = json.Unmarshal(msg.Data, &m)
				if err != nil {
					log.Println(err)
					return
				}
				if err = rs.Open(&m); err != nil {
					log.Println("open", err)
					return
				}
			case "read":
				d, err := rs.Read()
				if err != nil {
					if err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
						log.Println(err)
					}
				} else {
					if err := conn.WriteMessage(websocket.BinaryMessage, d); err != nil {
						log.Println(err)
					}
				}
			case "close":
				log.Println("close")
				return
			default:
				log.Println("unexpected msg", msg.Op)
				return
			}
		}
	}
}

func (u *Ui) apiGlob(w http.ResponseWriter, r *http.Request) {
	pattern := r.FormValue("pattern")

	rs := struct {
		Err  string
		Data []string
	}{}
	defer json.NewEncoder(w).Encode(&rs)

	c, err := findDial(u.store, 1)
	if err != nil {
		log.Println("find", err)
		return
	}

	defer func() {
		if e := c.Close(); e != nil {
			if rs.Err == "" {
				rs.Err = e.Error()
			}
		}
	}()

	files, err := c.Glob(pattern)
	if err != nil {
		rs.Err = err.Error()
		return
	}

	rs.Data = files
}
