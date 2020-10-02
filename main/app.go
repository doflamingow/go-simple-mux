package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gosimplemux/appHttpParser"
	"gosimplemux/appHttpResponse"
	"gosimplemux/appStatus"
	"log"
	"net/http"
)

type User struct {
	Id        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

var responder = appHttpResponse.NewJSONResponder()
var jsonParser = appHttpParser.NewJsonParser()
var users = []User{
	{
		Id:        "c01d7cf6-ec3f-47f0-9556-a5d6e9009a43",
		FirstName: "Edi",
		LastName:  "Uchida",
	},
}

func userRoute(w http.ResponseWriter, r *http.Request) {
	responder.Data(w, appStatus.Success, appStatus.StatusText(appStatus.Success), users)
}

func userPostRoute(w http.ResponseWriter, r *http.Request) {
	var newUser User
	if err := jsonParser.Parse(r, &newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if isAuth, ok := session.Values["authenticated"].(bool); !isAuth || !ok {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	id := guuid.New()
	newUser.Id = id.String()
	users = append(users, newUser)
	responder.Data(w, appStatus.Success, appStatus.StatusText(appStatus.Success), newUser)
}

func userPutRoute(w http.ResponseWriter, r *http.Request) {
	userId, isExist := r.URL.Query()["id"]
	var userUpdate User
	var userIdx int
	if isExist {
		for idx, usr := range users {
			if usr.Id == userId[0] {
				userUpdate = usr
				userIdx = idx
				break
			}
		}
		var usrReq User
		if err := jsonParser.Parse(r, &usrReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userUpdate.FirstName = usrReq.FirstName
		userUpdate.LastName = usrReq.LastName
		users[userIdx] = userUpdate
		responder.Data(w, appStatus.Success, appStatus.StatusText(appStatus.Success), userUpdate)
	} else {
		msg := appStatus.StatusText(appStatus.ErrorLackInfo)
		responder.Error(w, appStatus.ErrorLackInfo, fmt.Sprintf(msg, "ID"))
	}
}

func userDeleteRoute(w http.ResponseWriter, r *http.Request) {
	userId, isExist := r.URL.Query()["id"]
	var newUsers = make([]User, 0)
	if isExist {
		for _, usr := range users {
			if usr.Id == userId[0] {
				continue
			}
			newUsers = append(newUsers, usr)
		}
		users = newUsers
		responder.Data(w, appStatus.Success, appStatus.StatusText(appStatus.Success), nil)
	} else {
		msg := appStatus.StatusText(appStatus.ErrorLackInfo)
		responder.Error(w, appStatus.ErrorLackInfo, fmt.Sprintf(msg, "ID"))
	}

}
func authRoute(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "app-cookie")
	store.Options = &sessions.Options{
		MaxAge:   60 * 1,
		HttpOnly: true,
	}
	session.Values["authenticated"] = true
	err := session.Save(r, w)
	if err != nil {
		log.Fatalln(err)
	}
}
func authLogoutRoute(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "app-cookie")
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	fmt.Println(session.Values)
	if err != nil {
		log.Fatalln(err)
	}
}

var responder = appHttpResponse.NewJSONResponder()
var store = sessions.NewCookieStore([]byte("rahasia..."))

func main() {
	hostPtr := flag.String("host", "localhost", "Listening on host")
	portPtr := flag.String("port", "6969", "Listening on port")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/user", userRoute).Methods("GET")
	r.HandleFunc("/user", userPostRoute).Methods("POST")
	r.HandleFunc("/user", userPutRoute).Methods("PUT")
	r.HandleFunc("/user", userDeleteRoute).Methods("DELETE")
	r.HandleFunc("/user/{id}", userRoute).Methods("GET")

	r.HandleFunc("/login", authRoute)
	r.HandleFunc("/logout", authLogoutRoute)

	h := fmt.Sprintf("%s:%s", *hostPtr, *portPtr)
	log.Println("Listening on", h)
	err := http.ListenAndServe(h, r)
	if err != nil {
		log.Fatalln(err)
	}
}
