package main

import (
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"encoding/json"
	 "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "goji.io"
	"goji.io/pat"
	"fmt"
	"github.com/gorilla/securecookie"
)

// cookie handling

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func internalPageHandler(s *mgo.Session) func(res http.ResponseWriter, req *http.Request) {
	
	return func(res http.ResponseWriter, req *http.Request) {


	fmt.Println("Inside internalPageHandler")


	userName := getUserName(req)
	if userName != "" {
		fmt.Println(userName)

		u:=User{}

 		u, _ = FindUser(userName,s)

 		respondWithJson(res, 200, u)


	} else {
		http.Redirect(res, req, "/", 302)
	}
}

}


func setSession(userName string, response http.ResponseWriter) {


	fmt.Println("Inside setSession")


	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {

	fmt.Println("Inside clearSession")


	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}



// server main method


// MongoDB Config
var mongodb_server = "mongodb://54.67.121.40:27017,54.67.79.72:27017,54.193.98.71:27017,13.57.11.69:27017,13.57.59.232:27017/?replicaSet=rs0"
var mongodb_database = "cmpe281"
var mongodb_collection = "users"


type User struct {
    User_id string `bson:"user_id" json:"user_id"`
    UserName string `bson:"username" json:"username"`
    Password string `bson:"password" json:"password"`
}

var db *mgo.Database



// Select * from users where username=""
func FindUser(username string,s *mgo.Session) (User,error) {
 
    u := User{}


      
    session := s.Copy()
    defer session.Close()
 
    db=session.DB(mongodb_database)


    c := db.C(mongodb_collection)

    err := c.Find(bson.M{"username":username}).One(&u)
  
   if  err != nil {

   	// user with usrname exists

    return u, err
  }
   
   return u, nil

}


// Select password from users where username=""
func CheckPassword(username string,password string,s *mgo.Session) (User,error) {
 
    u := User{}

      
    session := s.Copy()
    defer session.Close()
 
    db=session.DB(mongodb_database)


    c := db.C(mongodb_collection)

    err := c.Find(bson.M{"username":username}).One(&u)
  
   if  err != nil {

   	// user with usrname exists

    return u, err
  }
   
   return u, nil

}


/*
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

*/




func signupPageHandler(s *mgo.Session) func(res http.ResponseWriter, req *http.Request) {
	
	return func(res http.ResponseWriter, req *http.Request) {


   fmt.Println("Inside signupPageHandler")

	if req.Method != "POST" {
		http.ServeFile(res, req, "UserForm/signup.html")
		return
	}

	// Decode credentials from json body
	u := &User{}

	err := json.NewDecoder(req.Body).Decode(u)

	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		res.WriteHeader(http.StatusBadRequest)
		return 
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 8)

		uid, _ := uuid.NewV4()
       
        u.User_id=uid.String()
	
		u.Password=string(hashedPassword)

	session := s.Copy()
     defer session.Close()

    c := session.DB(mongodb_database).C(mongodb_collection)


	if err := c.Insert(&u); err != nil {
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(res, http.StatusCreated, u)


}

}



func loginHandler(s *mgo.Session) func(res http.ResponseWriter, req *http.Request) {


	return func(res http.ResponseWriter, req *http.Request) {

  
   fmt.Println("Inside loginHandler")
 	
// Parse and decode the request body into a new `User` instance	
	usr := &User{}
	err := json.NewDecoder(req.Body).Decode(usr)

	if err != nil {
		// If there is something wrong with the request body, return a 400 status		
		respondWithError(res, http.StatusInternalServerError, err.Error())
		return 
	}

     session := s.Copy()
     defer session.Close()

    c := session.DB(mongodb_database).C(mongodb_collection)

	storedUsr := &User{}
	// Get the existing entry present in the database for the given username
	err = c.Find( bson.M{"username":usr.UserName}).Select(bson.M{"password": 1}).One(&storedUsr)


	if err != nil {
		respondWithError(res, http.StatusInternalServerError,"UserName doesnt exist")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUsr.Password), []byte(usr.Password))
	
	if err != nil {
		//http.Redirect(res, req, "UserForm/login.html", 301)
		respondWithError(res, http.StatusInternalServerError, "Invalid Password")
		return
	}


	u := User{}

    err = c.Find(bson.M{"username":usr.UserName}).One(&u)

   
    setSession(u.UserName, res)

	//res.Write([]byte("Hello " + usr.UserName))

	respondWithJson(res, http.StatusCreated, u)

}


}


func logoutHandler(s *mgo.Session) func(res http.ResponseWriter, req *http.Request) {
	
	return func(res http.ResponseWriter, req *http.Request) {

 	fmt.Println("Inside logoutHandler")

    clearSession(res)
	http.Redirect(res, req, "/", 302)


}

}



func homePageHandler(s *mgo.Session) func(res http.ResponseWriter, req *http.Request) {
	
	return func(res http.ResponseWriter, req *http.Request) {

		fmt.Println("Inside homePageHandler")
//	http.ServeFile(res, req, "UserForm/index.html")


}

}


func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(mongodb_database).C(mongodb_collection)

	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}


func pingHandler (s *mgo.Session) func(res http.ResponseWriter, req *http.Request) {
	
	return func(res http.ResponseWriter, req *http.Request) {
		
		respondWithJson(res, 200 , " Test Alive ")
	//	json.NewEncoder(w)
}

}


func main() {

	session, err := mgo.Dial(mongodb_server)
    
        if err != nil {
                panic(err)
        }

        defer session.Close()
        session.SetMode(mgo.Monotonic, true)

        ensureIndex(session)


     // Init router
  r := goji.NewMux()


// Route handles & endpoints
  //http.HandleFunc("/ping", pingHandler)

  r.HandleFunc(pat.Get("/ping"), pingHandler(session))
  
  r.HandleFunc(pat.Post("/login"), loginHandler(session))
  r.HandleFunc(pat.Post("/signup"), signupPageHandler(session))
  r.HandleFunc(pat.Post("/logout"), logoutHandler(session))
 
  r.HandleFunc(pat.Get("/"), homePageHandler(session)) // health check
  
  r.HandleFunc(pat.Get("/internal"), internalPageHandler(session))

 

  fmt.Println("Starbucks server listening on port 8000")

  http.ListenAndServe(":8000", r)


}


