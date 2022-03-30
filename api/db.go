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

	f "github.com/fauna/faunadb-go/v5/faunadb"
)

type LOC struct {
	City        string `json:"Name"`
	Country     string
	CountryCODE string `json:"Country"`
	CityCODE    string `json:"Location"`
	SubDIV      string `json:"Subdivision"`
	Coordinates string `json:"Coordinates"`
}

type Gist struct {
	URL string `json:"raw_url"`
}

type DATA map[string]f.Value
type rv f.RefV

func templ(id string) (string, error) {

	req, err := http.NewRequest("GET", "https://api.github.com/gists/"+id, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var raw Gist

	err = json.Unmarshal(body, &raw)
	if err != nil {
		return "", err
	}

	return raw.URL, nil

}

func Data(w http.ResponseWriter, r *http.Request) {

	var (
		data DATA
		rvs  []rv
	)

	resp, err := http.Get("https://gist.githubusercontent.com/ssskip/5a94bfcd2835bf1dea52/raw/3b2e5355eb49336f0c6bc0060c05d927c2d1e004/ISO3166-1.alpha2.json")
	if err != nil {
		fmt.Fprint(w, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprint(w, err)
	}

	country := make(map[string]string)

	err = json.Unmarshal(body, &country)
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

	switch r.Method {

	case "POST":

		resp, err := http.Get("https://raw.githubusercontent.com/ovrclk/un-locode/master/data/code-list_json.json")
		if err != nil {
			fmt.Fprint(w, err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}

		l := make([]LOC, 0)

		err = json.Unmarshal(body, &l)
		if err != nil {
			fmt.Fprint(w, err)
		}

		m := make([]LOC, 0)

		r.ParseForm()

		d := strings.ToUpper(strings.TrimSpace(r.FormValue("data")))

		for i := range l {

			if v, ok := country[l[i].CountryCODE]; ok {

				l[i].Country = v

			}

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

			if strings.Contains(city, "(") {

				n := strings.Split(strings.TrimSuffix(city, ")"), "(")

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

		url, err := templ("62bd80eaac41fb0251d87be53f804a4f")
		if err != nil {
			fmt.Fprint(w, err)
		}

		resp, err = http.Get(url)
		if err != nil {
			fmt.Fprint(w, err)
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}

		sort.SliceStable(m, func(i, j int) bool {
			return m[i].Country < m[j].Country
		})

		t, err := template.New("db").Parse(string(body))
		if err != nil {
			fmt.Fprint(w, err)
		}

		t.Execute(w, m)

	case "GET":

		s := make([]string, 0)

		for i := range rvs {

			if _, ok := country[rvs[i].ID]; ok {

				s = append(s, country[rvs[i].ID])

			}
		}

		resp, err = http.Get("https://gist.githubusercontent.com/mmaedel/e7f0c7f12fbd734a3e1f241503cf6915/raw/04cab3906d19cfd448c18a5e5c0bd6c95773008c/1.html")
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
