package clients

import (
    "net/http"
    "log"
    "fmt"

    "github.com/schmichael/whispering-gophers/util"
)

func Home(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "<h1>Send a message</h1>"+
        "<form action=\"/message/create/\" method=\"POST\">"+
        "<textarea name=\"body\"></textarea><br>"+
        "<input type=\"submit\" value=\"Save\">"+
        "</form>")
}

func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
    body := r.FormValue("body")
    m := Message{
        ID:   util.RandomID(),
        Addr: Self,
        Body: body,
        Nick: SelfNick,
    }
    broadcast(m)
    http.Redirect(w, r, "/", http.StatusFound)
}

func StartClient() {
    http.HandleFunc("/", Home)
    http.HandleFunc("/message/create/", SendMessageHandler)
    err := http.ListenAndServe(":12345", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}