package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	//"strconv"
	"time"

	f "github.com/fauna/faunadb-go/v5/faunadb"
)

type LOC struct {
	City        string `json:"Name"`
	CountryCODE string `json:"Country"`
	CityCODE    string `json:"Location"`
	SubDIV      string `json:"Subdivision"`
	Coordinates string `json:"Coordinates"`
}

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

		resp, err := http.Get("https://raw.githubusercontent.com/ovrclk/un-locode/master/data/code-list_json.json")
		if err != nil {
			fmt.Fprint(w, err)
		}

		time.Sleep(3e7)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}

		l := make([]LOC, 0)

		err = json.Unmarshal(body, &l)
		if err != nil {
			fmt.Fprint(w, err)
		}

		fmt.Fprint(w, l)

		m := make([]LOC, 0)

		r.ParseForm()

		d := strings.ToUpper(r.FormValue("data"))

		for i := range l {

			city := strings.ToUpper(l[i].City)

			if city == d {

				m = append(m, l[i])

				continue

			}

			if strings.Contains(city, "/") {

				n := strings.Split(city, "/")

				for j := range n {
					if n[j] == d {
						m = append(m, l[i])
						break
					}
				}

				continue

			}

			if strings.Contains(city, "-") {

				n := strings.Split(city, "-")

				for j := range n {
					if n[j] == d {
						m = append(m, l[i])
						break
					}
				}

				continue

			}

			n := strings.Fields(city)

			k := len(n)

			if k > 1 {

				for j := 0; j < k; j++ {

					if n[j] == d {

						m = append(m, l[i])

						continue

					}

				}

			}

		}

		resp, err = http.Get("https://gist.githubusercontent.com/mmaedel/00dbb8cc7416c8afe7b0ce441bc48a17/raw/cbca25d2bf333bd580a140226524546531a019ab/tmpl.html")
		if err != nil {
			fmt.Fprint(w, err)
		}
		//We Read the response body on the line below.
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}

		t, err := template.New("db").Parse(string(body))
		if err != nil {
			fmt.Fprint(w, err)
		}

		t.Execute(w, m)

	case "GET":

		s := make([]string, 0)

		for i := range rvs {

			if _, ok := l[rvs[i].ID]; ok {

				s = append(s, l[rvs[i].ID])

			}
		}

		resp, err = http.Get("https://gist.githubusercontent.com/mmaedel/00dbb8cc7416c8afe7b0ce441bc48a17/raw/cbca25d2bf333bd580a140226524546531a019ab/tmpl.html")
		if err != nil {
			fmt.Fprint(w, err)
		}
		//We Read the response body on the line below.
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}

		t, err := template.New("db").Parse(string(body))
		if err != nil {
			fmt.Fprint(w, err)
		}

		t.Execute(w, s)

	}

}
