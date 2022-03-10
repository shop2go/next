package handler

import (
	"fmt"
	"net/http"
	"os"
	//"time"

	f "github.com/fauna/faunadb-go/v5/faunadb"
)

type DATA map[string]f.Value
type rv f.RefV

func Handler(w http.ResponseWriter, r *http.Request) {

	var (
		data DATA
		rvs  []rv
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

	for i := range rvs {
		fmt.Fprint(w, rvs[i].ID)
	}
}
