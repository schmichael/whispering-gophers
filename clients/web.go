package clients

import (
    "net/http"
    "io"
    "log"
)

func Home(w http.ResponseWriter, req *http.Request) {
    io.WriteString(w, "hello, world!\n")
}

func StartClient() {
    http.HandleFunc("/", Home)
    err := http.ListenAndServe(":12345", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}