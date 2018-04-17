package main

import (
    "net/http"
    "personal_project1/models"
    "strconv"
    "encoding/json"
)

func main() {
    http.HandleFunc("/getdetail", getDetail)
    http.ListenAndServe(":4000", nil)
}

func getDetail(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    // in case wrong http method is used
    if r.Method != "GET" {
        w.Header().Set("Allow", "GET")
        http.Error(w, http.StatusText(405), 405)
        return
    }
    //in case id is empty
    id := r.URL.Query().Get("id")

    coffee, err := models.FindCoffee(id)
    if err == models.ErrNoProduct {
        http.NotFound(w, r)
        return
    } else if err != nil {
        http.Error(w, http.StatusText(500), 500)
        return
    }
    json.NewEncoder(w).Encode(coffee)
}