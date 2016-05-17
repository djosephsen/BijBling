package main

import(
	"fmt"
	"net/http"
	"net/url"
	"os"
   "github.com/Jeffail/gabs" // more easier json parsing
	"io/ioutil"
	"strings"
	)

func helpHandler(w http.ResponseWriter, r *http.Request) {
// Help screen for those accessing url '/'
   fmt.Fprintf(w, "Welcome to BijBling\nValid targets:\n\n")
   fmt.Fprintf(w, "/alert :: reflects websocket alerts from Librato\n")
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
// Handler for url path '/alert', this transforms an alert from librato
// into an annotation stream back inside librato
	luser := os.Getenv("BIJ_USER")
   lpass := os.Getenv("BIJ_TOKEN")

	//panic if we don't have librato credentials
	if luser == `` || lpass == `` {
		fmt.Printf("ERROR::Export BIJ_USER and BIJ_TOKEN with your librato credentials")
	}

	//parse out the server name and alert name from the alert payload
	r_form := r.ParseForm()
	payload := r_form.Get("payload")
	parsed_payload,_ := gabs.ParseJSON(payload)
	account,_ := parsed_payload.Path("account").Data().(string)
   alert, _ := parsed_payload.Path("alert.name").Data().(string)
	url_path := fmt.Sprintf("https://metrics-api.librato.com/v1/annotations/alerts.%s", alert)

	//Only librato.com accounts may use
	if strings.HasSuffix(account,"librato.com"){
		fmt.Printf("Dropping request from %s",account)
		return
	}

   violators, _ := parsed_payload.S("violations").ChildrenMap()
	for host, _ := range violators {
		//for each host that violated the threshold, fire off an annotation 
		//into the annotation stream named for this alert
		form := url.Values{}
		form.Set("title", fmt.Sprintf("Alert Fired :: %s ",alert))
		form.Set("source", host)
		form.Set("description", fmt.Sprintf("Alert: %s fired on host %s",alert, host))
		req,_ := http.NewRequest("POST", url_path, strings.NewReader(form.Encode()))
		defer req.Body.Close()
		req.SetBasicAuth(luser,lpass)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		fmt.Printf("Alert: %s fired on host %s\n", alert, host)
		//post the annotation 
		client := new(http.Client)
		resp, err := client.Do(req)

		//log the response from librato-api
		if err != nil {
			fmt.Printf("ERROR :: %s\n",err)
		}
		resp_body,_ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Annotation response: %s\n", resp.Status)
		fmt.Printf("Annotation response body: %s\n", resp_body)
	}
}

//lets kick this pig
func main() {
	port := fmt.Sprintf(":%s",os.Getenv("PORT"))
	fmt.Printf("Starting on port %s",port)
   http.HandleFunc("/", helpHandler)
   http.HandleFunc("/alert", alertHandler)
   http.ListenAndServe(port, nil)
}
