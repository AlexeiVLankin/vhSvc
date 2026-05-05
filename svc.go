package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

type tConnection struct {
	XMLName   xml.Name `xml:"Connection"`
	Name      string   `xml:"name,omitempty,attr"`
	DSN       string   `xml:"dsn,omitempty,attr"`
	URL       string   `xml:"url,omitempty,attr"`
	TestQuery string   `xml:"testQuery,omitempty,attr"`
}

type tSvcConfig struct {
	XMLName xml.Name `xml:"Preferences"`
	SVC     struct {
		Port            int `xml:"port,attr"`
		ShutdownTimeout int `xml:"shutdowntimeout,attr"`
	} `xml:"SVC"`
	GSProxy struct {
		URL string `xml:"url,omitempty,attr"`
	} `xml:"GSProxy"`
	Connections struct {
		Connection []tConnection
	} `xml:"Connections"`
}

const svcName = "Tracker Svc"

var done chan os.Signal
var restartAllowed bool

func checkServiceStatus(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, checkDBConnections())
	return
}

func checkRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Println(r.Header)
	respBody, _ := ioutil.ReadAll(r.Body)
	response := string(respBody)
	fmt.Println(response)
	io.WriteString(w, "test")
	return
}

type serviceCmd struct {
	signal string
}

func (c serviceCmd) String() string {
	return c.signal
}
func (c serviceCmd) Signal() {
}
func restartHTTPServer() {
	if !restartAllowed {
		return
	}
	restartAllowed = false
	done <- serviceCmd{"restart"}
}
func restartService(w http.ResponseWriter, r *http.Request) {
	restartHTTPServer()
	return
}

func stopService(w http.ResponseWriter, r *http.Request) {
	if !restartAllowed {
		return
	}
	restartAllowed = false
	go func() {
		done <- serviceCmd{"stop"}
	}()

	return
}

func svcCreateToken(w http.ResponseWriter, r *http.Request) {
	var token Token
	var callResult WebCallResult

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")

	err := xml.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		callResult.Result = "error"
		callResult.ErrorDesc = "error parsing token xml" + err.Error()
		fmt.Println(callResult)
		xml.NewEncoder(w).Encode(callResult)
		return
	}

	token.Status = ""
	err = createToken(&token, requestHeader(r))
	if err != nil {
		callResult.Result = "error"
		callResult.ErrorDesc = "could not create new token: " + err.Error()
		fmt.Println(callResult)
		xml.NewEncoder(w).Encode(callResult)
		return
	}

	callResult.Result = "OK"
	callResult.ErrorDesc = ""
	xml.NewEncoder(w).Encode(callResult)
	return
}

func getTokenInfoXML(w http.ResponseWriter, r *http.Request) {
	var callResult WebCallResult
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	vars := mux.Vars(r)
	id := vars["id"]
	pToken, err := getToken(id)
	if err != nil {
		callResult.Result = "error"
		callResult.ErrorDesc = "could not get token info"
		xml.NewEncoder(w).Encode(callResult)
		return
	}

	xml.NewEncoder(w).Encode(pToken)
	return
}

func setTokenInfoDialog(w http.ResponseWriter, r *http.Request) {
	const htmlTmpl = `<!DOCTYPE html>
<html>
    <head>
    <style>
    
    body {
      margin: 0;
      font-family: Arial, sans-serif;
      background-color: #3668c5;
      color: #333;
    
      height1: 100vh;
    }
    
.message {
      margin: 0;
      font-family: Arial, sans-serif;
     font-size: 2em;
      color: #CCCCFF;
    
   
      height1: 100vh;
    }
    
    .datasubmittedmsg {
      margin: 0;
      font-family: Arial, sans-serif;
     font-size: 3em;
      color: #CCCCFF;
      display: flex;
      align-items: center;
      justify-content: center;
      height1: 100vh;
    }

 .vin{
	border-bottom1: 1px solid #364043;
  color: #000;
  font-size:x-large;
  font-weight: 600;
  padding: 0.5em 1em;
  text-align1: left;
}   

 .vehdesc{
	border-bottom1: 1px solid #364043;
  color: #000;
  font-size:x-large;
  font-weight: 600;
  padding: 0.5em 1em;
  text-align1: left;
}   

    .content {
      background-color: white;
      padding: 2rem;
      border-radius: 12px;
      max-width1: 600px;
      width: 100%;
      text-align: center;
      box-shadow: 0 0 15px rgba(0, 0, 0, 0.1);
    }
    
    h1 {
      color: #3668c5;
      font-size: 2rem;
      margin-bottom: 1rem;
    }
    p {
      font-size: 1.1rem;
      margin-bottom: 1rem;
    }
    .logo {
      margin-top: 0rem;
    }
    .logo img {
      height: 60px;
    }





table.table1 {
  width1: 50vw;
  height1: 80vh;
  background: #012B39;
  border-radius: 0.25em;
  border-collapse: collapse;
  align:center;
  margin1: 1em;
}



th {
  border-bottom: 1px solid #364043;
  color: #E2B842;
  font-size: 0.85em;
  font-weight: 600;
  padding: 0.5em 1em;
  text-align1: left;
}

td {
  color: #fff;
  font-weight: 400;
  padding: 0.65em 1em;
  border-top: 3px solid #012B39;
}

.disabled td {
  color: #4F5F64;
}

tbody tr {
  transition: background 0.25s ease;
}

tbody
 tr:hover {
  background: #014055;
}



.button1 {
	height:70px;
    width:100%;
 padding: 6px 14px;
  font-family: -apple-system, BlinkMacSystemFont, 'Roboto', sans-serif;
 font-size: 3em;
  border-radius: 6px;
  border: none;
 background: #6E6D70;
  box-shadow: 0px 0.5px 1px rgba(0, 0, 0, 0.1), inset 0px 0.5px 0.5px rgba(255, 255, 255, 0.5), 0px 0px 0px 0.5px rgba(0, 0, 0, 0.12);
  color: #DFDEDF;
  user-select: none;
  -webkit-user-select: none;
  touch-action: manipulation;
  cursor: pointer;

  
}

.button1:focus {
  box-shadow: inset 0px 0.8px 0px -0.25px rgba(255, 255, 255, 0.2), 0px 0.5px 1px rgba(0, 0, 0, 0.1), 0px 0px 0px 3.5px rgba(58, 108, 217, 0.5);
  outline: 0;
}

input[type=checkbox]
{
  /* Double-sized Checkboxes */
  -ms-transform: scale(3); /* IE */
  -moz-transform: scale(3); /* FF */
  -webkit-transform: scale(3); /* Safari and Chrome */
  -o-transform: scale(3); /* Opera */
  transform: scale(3 );
  padding: 10px;
}

.make.abarth { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/abarth.gif); background-repeat: no-repeat; }
.make.alfa.romeo {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/alfa_romeo.gif); background-repeat: no-repeat; }
.make.audi {  border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/audi.gif); background-repeat: no-repeat; }
.make.bmw { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/bmw.gif); background-repeat: no-repeat; }
.make.chevrolet { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/chevrolet.gif); background-repeat: no-repeat; }
.make.chrysler { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/chrysler.gif); background-repeat: no-repeat; }
.make.citroen { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/citroen.gif); background-repeat: no-repeat; }
.make.cupra { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/cupra.gif); background-repeat: no-repeat; }
.make.dacia {  border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/dacia.gif); background-repeat: no-repeat;  }
.make.daihatsu {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/daihatsu.gif); background-repeat: no-repeat; }
.make.ds { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/ds.gif); background-repeat: no-repeat; }
.make.fiat { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/fiat.gif); background-repeat: no-repeat; }
.make.ferrari { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/ferrari.gif); background-repeat: no-repeat; }
.make.ford { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/ford.gif); background-repeat: no-repeat; }
.make.honda { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/honda.gif); background-repeat: no-repeat; }
.make.hyundai { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/hyundai.gif); background-repeat: no-repeat; }
.make.isuzu { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/isuzu.gif); background-repeat: no-repeat; }
.make.jaguar {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/jaguar.gif); background-repeat: no-repeat; }
.make.jeep { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/jeep.gif); background-repeat: no-repeat; }
.make.kia {  border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/kia.gif); background-repeat: no-repeat; }
.make.lancia { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/lancia.gif); background-repeat: no-repeat; }
.make.lada { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/lada.gif); background-repeat: no-repeat; }
.make.land.rover {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/land_rover.gif); background-repeat: no-repeat; }
.make.lexus { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/lexus.gif); background-repeat: no-repeat; }
.make.lotus { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/lotus.gif); background-repeat: no-repeat; }
.make.mazda { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/mazda.gif); background-repeat: no-repeat; }
.make.maserati {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/maserati.gif); background-repeat: no-repeat; }
.make.mercedes { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/mercedes.gif); background-repeat: no-repeat; }
.make.mg { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/mg.gif); background-repeat: no-repeat; }
.make.mini { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/mini.gif); background-repeat: no-repeat; }
.make.mitsubishi {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/mitsubishi.gif); background-repeat: no-repeat; }
.make.nissan {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/nissan.gif); background-repeat: no-repeat; }
.make.opel { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/opel.gif); background-repeat: no-repeat; }
.make.peugeot { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/peugeot.gif); background-repeat: no-repeat; }
.make.porsche { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/porsche.gif); background-repeat: no-repeat; }
.make.renault { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/renault.gif); background-repeat: no-repeat; }
/* NOUVEAU logo NEVS */ .make.saab {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/photos/logo2/saab.gif); background-repeat: no-repeat; }
.make.seat { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/seat.gif); background-repeat: no-repeat; }
.make.skoda { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/skoda.gif); background-repeat: no-repeat; }
.make.ssangyong { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/ssangyong.gif); background-repeat: no-repeat; }
.make.subaru { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/subaru.gif); background-repeat: no-repeat; }
.make.suzuki { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/suzuki.gif); background-repeat: no-repeat; }
.make.toyota { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/toyota.gif); background-repeat: no-repeat; }
.make.volkswagen { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/volkswagen.gif); background-repeat: no-repeat; }
.make.volvo { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/volvo.gif); background-repeat: no-repeat; }
.make.tesla, .make.gaz {border1: 3px solid #012B39; background-position:center;  background-image: url(https://secure.b-cars.ch/img/websites/int/icons/ico_question.svg); background-repeat: no-repeat; }
.make.jac { border1: 3px solid #012B39; background-position:center; background-image: url(https://secure.b-cars.ch/photos/logo2/jac.gif); background-repeat: no-repeat; }

    </style>
        <meta charset="UTF-8">
        <meta name="viewport" content=" initial-scale=0.5, maximum-scale=1, user-scalable=0"> 
       <title>Vehicles</title>
      
    </head>
    <body>
   
    <br><br><br>
        <form action="/processUpdate" method="POST">
           <input type="hidden" name="requestDetails" value={{.ID}} /> 
           <input type="hidden" name="id" value={{.ID}} /> 
               {{ if .Vehicles }}    
            <div align=top>  
        <table width="100%" class="table1">
     
         <tr>
         <td width=10% colspan=2> <div class="logo"><img src="../img/bcars_logo_white_nobg.png" alt="BCars Logo" /></div></td>
         <td width=85% colspan=2><div class="message">{{.EditMessage}} {{.GroupID}}:</div></td>
       
        </tr>
       
          {{ range .Vehicles }}
            <tr>          
            <td width=5% align=left  bgcolor=#cccccc class="make {{.Make}}"><div class="vehdesc">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</div></td>
           <td  width=10% bgcolor=#cccccc class="vehdesc">{{.Name}}</td>
            
           
             <td bgcolor=#cccccc width=10% align=left><input type="checkbox" name="{{.VIN}}"
           {{if $.Locked}} 
             disabled 
            {{end}}
           {{if .Loaded}} 
             checked 
            {{end}}
            /></td>
             <td bgcolor=#cccccc width=10% align=center><span class='vin'>{{.VIN}}</span></td>
            </tr>
         {{end}}
         
    <tr>
    <td colspan=2></td>
    <td colspan=2>    
    <div align=right>
   
    {{if .Locked}} 
            {{else}}
               <input type="submit" class="button1" value="Submit" />
            {{end}}
  
    </div>
    </td></tr>
    
      </table> 
    </div>   
     {{end}}           
        </form>
    </body>
</html>`

	tmpl, err := template.New("resp").Parse(string(htmlTmpl))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	pToken, err := getToken(id)

	if err != nil {
		io.WriteString(w, "Error getting token information")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/html; charset=utf-8")
	if err := tmpl.Execute(w, *pToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func setTokenInfoDialog1(w http.ResponseWriter, r *http.Request) {
	//device-width
	const htmlTmpl = `<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=200, initial-scale=1.0"> 
       <title>Vehicles</title>
        <link rel="stylesheet" type="text/css" href="../css/s1.css" />
    </head>
    <body>
   
    <br><br><br>
        <form action="/processUpdate" method="POST">
           <input type="hidden" name="id" value={{.ID}} /> 
               {{ if .Vehicles }}    
            <div align=center>  
        <table width="50%" class="table1">
     
         <tr>
         <td width=20%> <div class="logo"><img src="../img/bcars_logo_white_nobg.png" alt="BCars Logo" /></div></td>
         <td width=80% colspan=3><div class="message">{{.EditMessage}} {{.GroupID}}:</div></td>
       
        </tr>
       
          {{ range .Vehicles }}
            <tr>          
            <td width=70% align=left  bgcolor=#cccccc class="make {{.Make}}"><div class="vehdesc">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</div></td>
           <td  width=10% bgcolor=#cccccc class="vehdesc">{{.Name}}</td>
            
            <td bgcolor=#cccccc width=10% align=center><span class='vin'>{{.VIN}}</span></td>
             <td bgcolor=#cccccc width=10% align=left><input type="checkbox" name="{{.VIN}}"
           {{if $.Locked}} 
             disabled 
            {{end}}
           {{if .Loaded}} 
             checked 
            {{end}}
            /></td>
            </tr>
         {{end}}
         
    <tr><td colspan=4>    
    <div align=right>
   
    {{if .Locked}} 
            {{else}}
               <input type="submit" class="button1" value="Submit" />
            {{end}}
  
    </div>
    </td></tr>
      </table> 
    </div>   
     {{end}}           
        </form>
    </body>
</html>`

	tmpl, err := template.New("resp").Parse(string(htmlTmpl))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	pToken, err := getToken(id)

	if err != nil {
		io.WriteString(w, "Error getting token information")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/html; charset=utf-8")
	if err := tmpl.Execute(w, *pToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func dataSubmittedHTML(w http.ResponseWriter, r *http.Request) {
	const htmlTmpl = `<!DOCTYPE html>
<html>
    <head>
        <meta charset="UTF-8">
        <title>Vehicles</title>
        <link rel="stylesheet" type="text/css" href="../css/s1.css" />
    </head>
        <style>
    table.table1 {
  width: 100%;
  height: 50%;
  background: #012B39;
  border-radius: 0.25em;
  border-collapse: collapse;
  align:center;
  margin1: 1em;
}
    </style>
    <body>
   
    <br><br><br>
        <form action="/" method="GET">
           <input type="hidden" name="id" value={{.ID}} /> 
                 
        <table class="table1">
        <tr>
      
          <td align=left valign=top> <div class="logo"><img src="../img/bcars_logo_white_nobg.png" alt="BCars Logo" /></div></td>
        </tr>
        
        <tr> <td > <div class="datasubmittedmsg"> {{.Message2}}</div></td></tr>
        </table>
   
    </body>
</html>`

	tmpl, err := template.New("resp").Parse(string(htmlTmpl))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	pToken, err := getToken(id)

	if err != nil {
		io.WriteString(w, "Unknown token")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/html; charset=utf-8")
	if err := tmpl.Execute(w, *pToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func XML2Text(v interface{}) string {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(&v); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	return buf.String()
}

func requestHeader(r *http.Request) string {
	var requestDetails IncomingRequestDetails
	for name, values := range r.Header {
		for _, value := range values {
			requestDetails.Headers = append(requestDetails.Headers, HeaderNameValue{Name: name, Value: value})
		}
	}
	return XML2Text(requestDetails)
}

func updateTokenInfo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	id := r.FormValue("id")
	pToken, err := getToken(id)

	if err != nil {
		io.WriteString(w, "Token update error")
		return
	}

	for i := range pToken.Vehicles {
		vin := pToken.Vehicles[i].VIN
		pToken.Vehicles[i].Loaded = (r.FormValue(vin) == "on")
	}

	err = updateToken(pToken, []int{1, 2}, 2, "update", requestHeader(r)) //any status except "locked"
	if err != nil {
		io.WriteString(w, "Could not update token")
		return
	}

	http.Redirect(w, r, "/submitted/"+id, 301)
}

func svcUpdateToken(w http.ResponseWriter, r *http.Request) {
	var token Token
	var callResult WebCallResult

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")

	err := xml.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		callResult.Result = "error"
		callResult.ErrorDesc = "error parsing token xml" + err.Error()
		fmt.Println(callResult)
		xml.NewEncoder(w).Encode(callResult)
		return
	}

	err = updateToken(&token, []int{1}, 1, "local update", requestHeader(r)) //only new token can be updated by service
	if err != nil {
		callResult.Result = "error"
		callResult.ErrorDesc = err.Error()
		fmt.Println(callResult)
		xml.NewEncoder(w).Encode(callResult)
		return
	}

	callResult.Result = "OK"
	callResult.ErrorDesc = ""
	xml.NewEncoder(w).Encode(callResult)
	return

}

func svcLockToken(w http.ResponseWriter, r *http.Request) {
	var callResult WebCallResult

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	vars := mux.Vars(r)
	id := vars["id"]
	pToken, err := getToken(id)
	if err != nil {
		callResult.Result = "error"
		callResult.ErrorDesc = err.Error()
		xml.NewEncoder(w).Encode(callResult)
		return
	}

	err = updateToken(pToken, []int{1, 2}, 3, "lock", requestHeader(r)) //lock token
	if err != nil {
		callResult.Result = "error"
		callResult.ErrorDesc = err.Error()
		fmt.Println(callResult)
		xml.NewEncoder(w).Encode(callResult)
		return
	}

	callResult.Result = "OK"
	callResult.ErrorDesc = ""
	xml.NewEncoder(w).Encode(callResult)
	return

}

func startSVC() {

start:
	runtime.GC()
	restartAllowed = true
	r := mux.NewRouter()

	s := http.StripPrefix("/img/", http.FileServer(http.Dir("./img/")))
	r.PathPrefix("/img/").Handler(s)

	s = http.StripPrefix("/css/", http.FileServer(http.Dir("./css/")))
	r.PathPrefix("/css/").Handler(s)
	http.Handle("/", r)

	r.HandleFunc("/status", checkServiceStatus).Methods("GET")
	r.HandleFunc("/restart", restartService).Methods("GET")
	r.HandleFunc("/stop", stopService).Methods("GET")

	r.HandleFunc("/createtoken", svcCreateToken).Methods("POST")
	r.HandleFunc("/updatetoken", svcUpdateToken).Methods("POST")
	r.HandleFunc("/gettoken/{id}", getTokenInfoXML).Methods("GET")
	r.HandleFunc("/locktoken/{id}", svcLockToken).Methods("GET")
	r.HandleFunc("/submitted/{id}", dataSubmittedHTML).Methods("GET")
	r.HandleFunc("/update/{id}", setTokenInfoDialog).Methods("GET")

	r.HandleFunc("/processUpdate", updateTokenInfo).Methods("POST")

	sPort := strconv.Itoa(svcConfig.SVC.Port)

	srv := &http.Server{
		Addr:         ":" + sPort,
		Handler:      r,
		ReadTimeout:  10000 * time.Second,
		WriteTimeout: 10000 * time.Second,
	}

	log.Printf(svcName+" started on port: %v", sPort)

	done = make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		runtime.GC()
		err := srv.ListenAndServe()
		runtime.GC()
		if err != nil {
			log.Printf("HTTP Server: %v", err)
			done <- nil //error might be due to requested shutdown.
		}
	}()

	cmd := <-done
	log.Printf("Command received: %v. Stopping http server, timeout: %v sec", cmd, svcConfig.SVC.ShutdownTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(svcConfig.SVC.ShutdownTimeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Printf(svcName + " stopped!")
	switch cmd.String() {
	case "restart":
		//	log.Printf("Restarting VSpecProxy Service...")
		goto start
	case "stop":
		log.Printf(svcName + " disabled!")
	}
	return
}
