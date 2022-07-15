package client

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"logViewer/find/tcp"

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

	_, input, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	var m tcp.Match

	err = json.Unmarshal(input, &m)
	if err != nil {
		log.Println(err)
		return
	}

	addr := "127.0.0.2:7000"
	rs, err := tcp.Dial(addr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer rs.Close()

	//log.Printf("match %#v\n", &m)

	err = rs.Open(&m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for {
		conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))

		d, err := rs.Read()
		log.Println(string(d), err)
		if err != nil {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
				log.Println(2, err)
			}
			break
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, d); err != nil {
			log.Println(3, err)
			break
		}

		_, _, err = conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage", err)
			return
		}
	}

	time.Sleep(time.Second)
}

func (u *Ui) apiGlob(w http.ResponseWriter, r *http.Request) {
	pattern := r.FormValue("pattern")

	rs := struct {
		Err  string
		Data []string
	}{}
	defer json.NewEncoder(w).Encode(&rs)

	addr := "127.0.0.2:7000"
	c, err := tcp.Dial(addr)
	if err != nil {
		rs.Err = err.Error()
		return
	}
	defer c.Close()

	files, err := c.Glob(pattern)
	if err != nil {
		rs.Err = err.Error()
		return
	}

	rs.Data = files

}
