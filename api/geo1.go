package data

import (
	"fmt"
	"io"
	"net/http"
)

func Geo1(w http.ResponseWriter, r *http.Request) {

	resp, err := http.Get("https://vercel-geoip.vercel.app")
	if err != nil {
		fmt.Fprint(w, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprint(w, err)
	}

	fmt.Fprint(w, body)

}
