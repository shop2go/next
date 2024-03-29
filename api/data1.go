package data

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	o "golang.org/x/oauth2"

	f "github.com/fauna/faunadb-go/v5/faunadb"

	g "github.com/shurcooL/graphql"
)

type ACCESS struct {
	//Reference *f.RefV `fauna:"ref"`
	Timestamp int    `fauna:"ts"`
	Secret    string `fauna:"secret"`
	Role      string `fauna:"role"`
}

type LOC struct {
	City        string `json:"Name"`
	Country     string
	CountryCODE string `json:"Country"`
	CityCODE    string `json:"Location"`
	SubDIV      string `json:"Subdivision"`
	Coordinates string `json:"Coordinates"`
}

type LOCK struct {
	Link g.ID     `graphql:"link"`
	Data g.String `graphql:"data"`
}

type DATA map[string]f.Value
type rv f.RefV

type GIST struct {
	Files FILES `json:"files"`
}

type FILES struct {
	File1 RAW `json:"1.html"`
	File2 RAW `json:"2.html"`
}

type RAW struct {
	Raw string `json:"raw_url"`
}

func templ(id string) (GIST, error) {

	req, err := http.NewRequest("GET", "https://api.github.com/gists/"+id, nil)
	if err != nil {
		return GIST{}, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GIST{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GIST{}, err
	}

	var gist GIST

	err = json.Unmarshal(body, &gist)
	if err != nil {
		return GIST{}, err
	}

	return gist, nil

}

func Data1(w http.ResponseWriter, r *http.Request) {

	id := r.Header.Get("x-vercel-ip-country")

	fmt.Fprint(w, r.Header)

	country := make(map[string]string)

	resp, err := http.Get("https://gist.githubusercontent.com/ssskip/5a94bfcd2835bf1dea52/raw/3b2e5355eb49336f0c6bc0060c05d927c2d1e004/ISO3166-1.alpha2.json")
	if err != nil {
		fmt.Fprint(w, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprint(w, err)
	}

	err = json.Unmarshal(body, &country)
	if err != nil {
		fmt.Fprint(w, err)
	}

	switch r.Method {

	case "POST":

		resp, err = http.Get("https://raw.githubusercontent.com/ovrclk/un-locode/master/data/code-list_json.json")
		if err != nil {
			fmt.Fprint(w, err)
		}

		body, err = io.ReadAll(resp.Body)
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

		gist, err := templ(os.Getenv("GIST_ID"))
		if err != nil {
			fmt.Fprint(w, err)
		}

		resp, err = http.Get(gist.Files.File2.Raw)
		if err != nil {
			fmt.Fprint(w, err)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}

		sort.SliceStable(m, func(i, j int) bool {
			return m[i].Country < m[j].Country
		})

		t, err := template.New("data").Parse(string(body))
		if err != nil {
			fmt.Fprint(w, err)
		}

		t.Execute(w, m)

	case "GET":

		var (
			data DATA
			rvs  []rv
		)

		gist, err := templ(os.Getenv("GIST_ID"))
		if err != nil {
			fmt.Fprint(w, err)
		}

		resp, err := http.Get(gist.Files.File1.Raw)
		if err != nil {
			fmt.Fprint(w, err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}

		t, err := template.New("data").Parse(string(body))
		if err != nil {
			fmt.Fprint(w, err)
		}

		ep := f.Endpoint("https://db.fauna.com:443")

		fdb := os.Getenv("FAUNA_DB")

		c := f.NewFaunaClient(fdb, ep)

		x, err := c.Query(f.CreateKey(f.Obj{"database": f.Database("access"), "role": "server-readonly"}))
		if err != nil {
			fmt.Fprint(w, err)
		}

		var acc ACCESS

		x.Get(&acc)

		s := make([]string, 0)

		if id != "" {

			src := o.StaticTokenSource(
				&o.Token{AccessToken: acc.Secret},
			)

			httpClient := o.NewClient(context.Background(), src)

			call := g.NewClient("https://graphql.fauna.com/graphql", httpClient)

			var q struct {
				LOCKS struct {
					Data []LOCK
				} `graphql:"locks(data: $data)"`
			}
			vars := map[string]interface{}{
				"data": g.String(id),
			}

			if err := call.Query(context.Background(), &q, vars); err != nil {
				fmt.Fprint(w, err)
			}

			l := q.LOCKS.Data

			if l != nil {

				for _, v := range l {

					s = append(s, (v.Link).(string))

				}

				t.Execute(w, s)

			} else {

				d := f.NewFaunaClient(acc.Secret, ep)

				x, err := d.Query(f.Paginate(f.Documents(f.ScopedCollection("LOCK", f.Database(strings.ToUpper(id)))), f.Size(30)))
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

				for _, v := range rvs {

					var q struct {
						findLOCKByID struct {
							LOCK
						} `graphql:"findLOCKByID(id: $id)"`
					}

					vars := map[string]interface{}{
						"id": g.String(v.ID),
					}

					if err := call.Query(context.Background(), &q, vars); err != nil {
						fmt.Fprint(w, err)
					}

					s = append(s, string(q.findLOCKByID.Data))

				}

				t.Execute(w, s)

			}

		} else {

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

			s := make([]string, 0)

			for i := range rvs {

				if v, ok := country[rvs[i].ID]; ok {

					s = append(s, v)

				}
			}

			t.Execute(w, s)

		}

	}

}
