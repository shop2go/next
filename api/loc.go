package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"log"
	"net/http"
)

type LOC struct {
	City        string `json:"Name",omitempty`
	Country     string
	CountryCODE string `json:"Country",omitempty`
	CityCODE    string `json:"Location",omitempty`
	SubDIV      string `json:"Subdivision",omitempty`
	Coordinates string `json:"Coordinates",omitempty`
}

func Loc(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "https://loc.code2go.dev/data", http.StatusFound)

	resp, err := http.Get("https://raw.githubusercontent.com/ovrclk/un-locode/master/data/code-list_json.json")
	if err != nil {
		fmt.Fprint(w, err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprint(w, err)
	}

	var l []LOC

	err = json.Unmarshal(body, &l)
	if err != nil {
		fmt.Fprint(w, err)
	}

	for i := range l {

		if l[i].City == "Salzburg" {
			fmt.Fprint(w, l[i])
		}
	}
}
