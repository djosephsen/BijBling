package main

import(
	"fmt"
	"net/http"
	"net/url"
	"os"
   "github.com/Jeffail/gabs"
	"io/ioutil"
	"strings"
	)

func helpHandler(w http.ResponseWriter, r *http.Request) {
   fmt.Fprintf(w, "Welcome to BijBling\nValid targets:\n\n")
   fmt.Fprintf(w, "/alert :: reflects websocket alerts from Librato\n")
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
	luser := os.Getenv("BIJ_USER")
   lpass := os.Getenv("BIJ_TOKEN")

	payload,_ := ioutil.ReadAll(r.Body)
	parsed_payload,_ := gabs.ParseJSON(payload)
   alert, _ := parsed_payload.Path("alert.name").Data().(string)
	url_path := fmt.Sprintf("http://metrics-api.librato.com/v1/annotations/alerts.%s", alert)

   violators, _ := parsed_payload.S("violations").ChildrenMap()
	for host, _ := range violators {
		form := new(url.Values)
		form.Add("title", fmt.Sprintf("Alert Fired :: %s ",alert))
		form.Add("source", host)
		form.Add("description", fmt.Sprintf("Alert: %s fired on host %s",alert, host))
		req,_ := http.NewRequest("POST", url_path, strings.NewReader(form.Encode()))
		defer req.Body.Close()
		req.SetBasicAuth(luser,lpass)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		fmt.Printf("Alert: %s fired on host %s", alert, host)

		client := new(http.Client)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("ERROR :: %s\n",err)
		}
		fmt.Printf("Annotation response: %s", resp.Status)
	}
}

func main() {
   http.HandleFunc("/", helpHandler)
   http.HandleFunc("/alert/", alertHandler)
   http.ListenAndServe(":8080", nil)
}
