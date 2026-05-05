package main

import (
	"fmt"
	"strings"
)

const htmlDoc = `<html><style>
.tableFixHead          { overflow: auto; height: 95%; } 
.tableFixHead thead th { position: sticky; top: 0; z-index: 1; }
UL {margin-top:7px;margin-bottom:5px;font-family:Verdana; font-size:12px;}
BODY, HREF, TH, TD {font-family:Verdana; font-size:18px;}
</style>`

func zero2nbsp(s interface{}) string {
	rez := fmt.Sprintf("%v", s)
	if rez == "0" {
		rez = "&nbsp;"
	}
	return rez
}
func htmlStyle(s string, attr ...string) string {
	clr, stl := "", ""
	if len(attr) >= 1 {
		clr = "color=\"" + attr[0] + "\""
	}
	rez := "<font " + clr + "/>" + s + "</font>"
	if len(attr) >= 2 {
		stl = attr[1]
		if stl != "" {
			rez = "<" + stl + ">" + rez + "</" + stl + ">"
		}
	}
	return rez
}
func textHTML(pattern string) func(data interface{}) string {
	arr := strings.Split(pattern, "###")
	return func(data interface{}) string {
		return arr[0] + fmt.Sprintf("%v", data) + arr[1]
	}
}

func htmlBold(s interface{}) string {
	return "<B>" + fmt.Sprintf("%v", s) + "</B>"
}
func startTABLE(caption string, headers ...interface{}) string {
	rez := "" // "<style>BODY, TH, TD { font-size:18px; font-family:Verdana} BODY {color:black} </style>"
	rez += "<TABLE width='100%' Cellspacing='0' Border='1' bordercolor='#aaaaaa' >\r\n"
	rez += "<caption>" + caption + "</caption>"
	rez += "<thead><TR>"
	for _, s := range headers {
		rez += "<TH bgcolor=#DDDDDD>" + fmt.Sprintf("%v", s) + "</TH>"
	}
	rez += "</TR></thead>\r\n"
	return rez
}

func htmlTableStart(headers ...interface{}) string {
	rez := "<TABLE width=100% Cellspacing='0' Border='1' bordercolor='#dddddd' >\r\n"
	rez += "<thead><TR align=center>"
	for _, s := range headers {
		rez += "<TH bgcolor=#DDDDDD>" + fmt.Sprintf("%v", s) + "</TD>"
	}
	rez += "</TH></thead>\r\n<tbody>"
	return rez
}

func htmlTableEnd() string {
	return "</tbody></TABLE>"
}

func tdHTML(attr ...string) func(data interface{}) string {
	tdAttr := ""
	if len(attr) == 1 {
		tdAttr = " " + attr[0]
	}
	return func(data interface{}) string {
		return "<TD" + tdAttr + ">" + fmt.Sprintf("%v", data) + "</TD>"
	}
}

func tdData(data ...interface{}) string {
	tdAttr := ""
	for i, attr := range data {
		if i >= len(data)-1 { //last == data
			break
		}
		tdAttr += " " + fmt.Sprintf("%v", attr) + " "
	}
	return "<TD" + tdAttr + ">" + fmt.Sprintf("%v", data[len(data)-1]) + "</TD>"
}

func toText(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
func trHTML(attr ...string) func(td ...string) string {
	trAttr := ""
	if len(attr) == 1 {
		trAttr = " " + attr[0]
	}
	return func(td ...string) string {
		rez := "<TR" + trAttr + ">"
		for _, s := range td {
			rez += s
		}
		rez += "</TR>"
		return rez
	}
}

func hrefHTML(desc interface{}, link ...interface{}) string {
	sDesc := fmt.Sprintf("%v", desc)
	sLink := ""
	for _, s := range link {
		if sLink != "" {
			sLink += "/"
		}
		sLink += fmt.Sprintf("%v", s)
	}
	return "<a href='/" + sLink + "'>" + sDesc + "</a>"
}

func TR(strs ...interface{}) string {
	rez := "<TR>"
	for _, s := range strs {
		rez += "<TD>" + fmt.Sprintf("%v", s) + "</TD>"
	}
	rez += "</TR>\r\n"
	return rez
}

func onClick(params ...interface{}) string {
	//onClick="javascript:window.location.href='#$LoadMakeModel#DACIA#SANDERO#'"
	rez := " onClick=\"javascript:window.location.href='#"
	for _, p := range params {
		rez += "#" + fmt.Sprintf("%v", p)
	}
	rez += "'\" "
	return rez
}
func endTABLE() string {
	return "</TABLE></BODY>"
}

// func equipIDListHTML(w http.ResponseWriter) {

// 	fmt.Fprint(w, htmlDoc)
// 	txtEQ := textHTML("<span style='color:black;font-weight:bold1;'>###</span>")
// 	txtEQImplemented := textHTML("<span style='color:blue;font-weight:bold1;'>###</span>")
// 	txtEQCheckImplemented := textHTML("<span style='color:blue;background-color:#fff791;font-weight:bold1;'>###</span>")
// 	txtEQGenRequired := textHTML("<span style='color:red;font-weight:bold1;'>###</span>")
// 	txtEQMissing := textHTML("<span style='color:maroon;font-weight:bold1;'>###</span>")
// 	fmt.Fprint(w, "<div class='tableFixHead'>")
// 	fmt.Fprint(w, startTABLE("", "EquipID", "Equip. count", "Attr. count", "Equipment name" /*, "AV Stat Table", "AV Stat XML", "Generation Template"*/))
// 	td := tdHTML()
// 	tdr := tdHTML("align=right")
// 	tr := trHTML()
// 	for i := range equipIDindex {
// 		equipID := equipIDindex[i]
// 		if equipIDstatus[equipID] != 1 {
// 			continue
// 		}
// 		equipIDCnt := len(equipIDHashCount[equipID])
// 		eq := firstLetter2UpperCase(schemaML["fr"][equipID][""])
// 		s := txtEQ(eq)
// 		if equipIDImplemented(equipID) == 1 {
// 			s = txtEQImplemented(eq)
// 			if _, ok := importantValues[equipID]; ok {
// 				s = txtEQCheckImplemented(eq)
// 			}
// 		}
// 		if avStatXMLsMap[equipID].genRequired {
// 			s = txtEQGenRequired(eq)
// 		}
// 		if equipIDstatus[equipID] != 1 {
// 			s = txtEQMissing(eq)
// 		}
// 		ref := "<a href='/equipmentlist/fr/" + strconv.Itoa(equipID) + "/1000'\">"
// 		//	statXMLRef := "<a href='/avstat/" + strconv.Itoa(equipID) + "'\">"
// 		//	statTableRef := "<a href='/avstathtml/" + strconv.Itoa(equipID) + "'\">"
// 		//	templateRef := "<a href='/gentemplate/" + strconv.Itoa(equipID) + "'\">"
// 		fmt.Fprint(w, tr(tdr(ref+strconv.Itoa(equipID)), tdr(equipIDCnt), tdr(avStatXMLsMap[equipID].attrCount), td(ref+s) /*, td(statTableRef+"html"), td(statXMLRef+"xml"), td(templateRef+"genTemplate")*/))
// 	}
// 	fmt.Fprint(w, endTABLE())
// 	fmt.Fprint(w, "</div>")
// 	return
// }
