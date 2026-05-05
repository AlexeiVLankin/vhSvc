package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type Vehicle struct {
	VIN    string `xml:"VIN,attr"`
	Name   string `xml:"name,attr"`
	Make   string `xml:"make,attr"`
	Loaded bool   `xml:"loaded,attr"`
}

type Token struct {
	XMLName     xml.Name  `xml:"PickupToken"`
	ID          string    `xml:"ID,omitempty,attr"`
	GroupID     string    `xml:"groupid,omitempty,attr"`
	Status      string    `xml:"status,omitempty,attr"`
	Message1    string    `xml:"message1,omitempty,attr"`
	Message2    string    `xml:"message2,omitempty,attr"`
	Message3    string    `xml:"message3,omitempty,attr"`
	Vehicles    []Vehicle `xml:"Vehicles>Vehicle"`
	EditMessage string
	Locked      bool
}

type WebCallResult struct {
	XMLName   xml.Name `xml:"Call"`
	Result    string   `xml:"result,omitempty,attr"`
	Error     string   `xml:"error,omitempty,attr"`
	ErrorDesc string   `xml:"errorDesc,omitempty,attr"`
}

type HeaderNameValue struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}
type IncomingRequestDetails struct {
	XMLName xml.Name          `xml:"RequestInfo"`
	Headers []HeaderNameValue `xml:"Headers>Header"`
}

const testXML1 = `<PickupToken ID="f1dd97a-84ca-7847-9653-42db0b1be59e6" groupid="G1" displayName="T06473" status="New">
<Vehicles>
<Vehicle VIN="UU1DJF01274857299" name="DACIA DUSTER 1.2 MHEV TCE 130HP JOURNEY 4WD" loaded="true"/>
<Vehicle VIN="UU1DJF01074857298" name="DACIA DUSTER 1.2 MHEV TCE 130HP JOURNEY 4WD" loaded="true"/>
<Vehicle VIN="UU1DJF01171811710" name="DACIA DUSTER 1.2 MHEV TCE 130HP JOURNEY 4WD" loaded="true"/>
<Vehicle VIN="UU1DJF01374223919" name="DACIA DUSTER 1.2 MHEV TCE 130HP EXPRESSION 4WD" loaded="true"/>
<Vehicle VIN="UU1DJF01773736189" name="DACIA DUSTER 1.2 MHEV TCE 130HP EXTREME 4WD" loaded="true"/>
<Vehicle VIN="UU1DJF01174857276" name="DACIA DUSTER 1.2 MHEV TCE 130HP JOURNEY 4WD" loaded="true"/>
<Vehicle VIN="UU1DJF01974857249" name="DACIA DUSTER 1.2 MHEV TCE 130HP JOURNEY 4WD" loaded="true"/>
<Vehicle VIN="UU1DJF01X74857244" name="DACIA DUSTER 1.2 MHEV TCE 130HP JOURNEY 4WD" loaded="true"/>
</Vehicles>
</PickupToken>`

func testStructure() {
	byteValue := []byte(testXML1)
	var token Token
	xml.Unmarshal(byteValue, &token)
}

func XML2String(v interface{}) string {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("  ", "    ")
	if err := enc.Encode(&v); err != nil {
		fmt.Printf("error: %v\n", err)
		return ""
	}
	return buf.String()
}
