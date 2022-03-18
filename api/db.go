package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	//"html/template"
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

	for {
		//http.Redirect(w, r, "http://code2go.dev/data", http.StatusFound)

		switch r.Method {

		case "GET":

			str := `
	<!DOCTYPE html>
	<html lang="en">
		 <head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<meta http-equiv="X-UA-Compatible" content="ie=edge">
				<title>CODE2GO</title>
				<!-- Font Awesome -->
<link
  href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.1/css/all.min.css"
  rel="stylesheet"
/>
<!-- Google Fonts -->
<link
  href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,700&display=swap"
  rel="stylesheet"
/>
<!-- MDB -->
<link
  href="https://cdnjs.cloudflare.com/ajax/libs/mdb-ui-kit/3.10.2/mdb.min.css"
  rel="stylesheet"
/>
				</head>
				<body style="background-color: #bcbcbc;">
				<br>
				<div class="container-sm" id="data" style="color:white; font-size:30px;">
				<div class="form-outline mb-4" method= "POST">
				<div class="input-group">
				<button type="submit" class="btn btn-outline-primary">search</button>
				<input type="search" name="city" class="form-control rounded" placeholder="city or select country below" aria-label="Search" aria-describedby="search-addon" />
				</div
				</div><br><br>

					   <ul class="list-group">
	`

			for i := range rvs {

				if _, ok := l[rvs[i].ID]; ok {
					str = str + `
		<br><li class="list-group-item">
		` + l[rvs[i].ID] +
						`</li>`

				}
			}

			str = str + `
	</ul>
							  </div>
							  <script
							  type="text/javascript"
							  src="https://cdnjs.cloudflare.com/ajax/libs/mdb-ui-kit/3.10.2/mdb.min.js"></script>
							  </body>
							  </html>`

			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("Content-Length", strconv.Itoa(len(str)))
			w.Write([]byte(str))

		case "POST":

			r.ParseForm()

			fmt.Fprint(w, r.FormValue("city"))

		}

	}

}
