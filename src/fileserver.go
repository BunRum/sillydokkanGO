package misc

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var FileServerUrl string
var FileServer *http.Server

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(fmt.Sprintf("assets/%s", r.URL.Path)); err == nil {

		} else {
			log.Printf("%s %s does not exist", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

func StartFileServer() {
	//srv := &http.Server{Addr: ":8080"}
	Settings := GetSettings()
	FileServer := &http.Server{Addr: ":8082", Handler: logRequests(http.FileServer(http.Dir(Settings["AssetPath"].(string))))}
	// always returns error. ErrServerClosed on graceful close
	err := FileServer.ListenAndServeTLS("./server.crt", "./server.key")
	if err != nil {
		return
	}
}
func RestartFileServer() {
	if FileServer != nil {
		err := FileServer.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}
	StartFileServer()
}
