package handler

import (
	"fmt"
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

	//http.Redirect(w, r, "http://code2go.dev/data", http.StatusFound)

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
					   <div class="container" id="data" style="color:white; font-size:30px;">
					   <ul class="list-group">
	`

	for i := range rvs {

		str = str + `
		<br><li class="list-group-item">
		` + rvs[i].ID +
			`</li>`
	}

	str = str + `
	</ul>
							  </div>
							  <script
							  type="text/javascript"
							  src="https://cdnjs.cloudflare.com/ajax/libs/mdb-ui-kit/3.10.2/mdb.min.js"
							></script>
							  </body>
							  </html>`

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(str)))
	w.Write([]byte(str))

}
