package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gocraft/web"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	uuid     string
	username string
}

type Context struct {
	Token    string
	UserInfo map[string]interface{}
}

func (c *Context) UserRequired(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	var userInfo map[string]interface{}
	if req.URL.Path == "/healthcheck" {
		next(rw, req)
		return
	}

	if req.URL.Path == "/signup" {
		next(rw, req)
		return
	}

	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := auth[7:]
	client := &http.Client{}
	authReq, err := http.NewRequest("GET", "https://meshblu.octoblu.com/v2/whoami", nil)
	authReq.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(authReq)
	if err != nil {
		panic(err)
		return
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		panic(err)
	}

	c.Token = token
	c.UserInfo = userInfo
	next(rw, req)
}

func (c *Context) GetUserInfo(rw web.ResponseWriter, req *web.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	session, err := mgo.Dial(os.Getenv("MONGODB_URI"))
	if err != nil {
		panic(err)
		return
	}
	collection := session.DB("crossyio-main").C("users")
	query := make(map[string]interface{})
	query["uuid"] = c.UserInfo["uuid"]
	result := User{}
	merr := collection.Find(query).One(&result)
	if merr != nil {
		panic(merr)
		return
	}
	fmt.Println(result)
	if err := json.NewEncoder(rw).Encode(result); err != nil {
		panic(err)
	}
}

func (c *Context) Healthcheck(rw web.ResponseWriter, req *web.Request) {
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintf(rw, "{\"online\": true}")
}

func (c *Context) RedirectOauth(rw web.ResponseWriter, req *web.Request) {
	rw.Header().Set("Location", "https://oauth.crossy.io/?callbackUrl=https://info.crossy.io/oauth/callback")
	rw.WriteHeader(http.StatusFound)
}

func main() {
	router := web.New(Context{}). // Create your router
					Middleware(web.LoggerMiddleware).     // Use some included middleware
					Middleware(web.ShowErrorsMiddleware). // ...
					Middleware((*Context).UserRequired).
					Get("/healthcheck", (*Context).Healthcheck).
					Get("/api/v1/users/:uuid", (*Context).GetUserInfo). // Add a route
					Get("/signup", (*Context).RedirectOauth)            // Add a route
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	http.ListenAndServe("0.0.0.0:"+port, router) // Start the server!
}
