package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	//"strconv"

	//"time"

	f "github.com/fauna/faunadb-go/v5/faunadb"
)

type DATA map[string]f.Value
type rv f.RefV

func DB(w http.ResponseWriter, r *http.Request) {

	var (
		data DATA
		rvs  []rv
		//str  string
	)

	resp, err := http.Get("https://gist.githubusercontent.com/ssskip/5a94bfcd2835bf1dea52/raw/3b2e5355eb49336f0c6bc0060c05d927c2d1e004/ISO3166-1.alpha2.json")
	if err != nil {
		fmt.Fprint(w, err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprint(w, err)
	}

	l := make(map[string]string)

	err = json.Unmarshal(body, &l)
	if err != nil {
		fmt.Fprint(w, err)
	}

	ep := f.Endpoint("https://db.fauna.com:443")

	fdb := os.Getenv("FAUNA_DB")

	c := f.NewFaunaClient(fdb, ep)

	x, err := c.Query(f.Paginate(f.Databases()))

	if err != nil {
		fmt.Fprint(w, err)
	}

	//log.Println(x)

	if err = x.Get(&data); err != nil {
		fmt.Fprint(w, err)
	}

	x = data["data"]

	if err = x.Get(&rvs); err != nil {
		fmt.Fprint(w, err)
	}

	sort.SliceStable(rvs, func(i, j int) bool {
		return rvs[i].ID < rvs[j].ID
	})

	//http.Redirect(w, r, "http://code2go.dev/data", http.StatusFound)

	switch r.Method {

	case "POST":

		r.ParseForm()

		fmt.Fprint(w, r.FormValue("city"))

	case "GET":

		t := template.New("db")
		t, _ = t.ParseFiles("/public/tmpl.html")
		t.Execute(w, rvs)

	}

}
