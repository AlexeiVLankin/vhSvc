package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dataField = struct {
	XMLName xml.Name `xml:"Field"`
	Value   string   `xml:",chardata"`
	Name    string   `xml:"name,omitempty,attr"`
}

type dataRow struct {
	XMLName xml.Name    `xml:"Row"`
	Index   int         `xml:"index,attr"`
	Field   []dataField `xml:"Field"`
}
type queryParam struct {
	XMLName xml.Name `xml:"Param"`
	Name    string   `xml:"name,omitempty,attr"`
	Type    string   `xml:"type,omitempty,attr"`
	Format  string   `xml:"format,omitempty,attr"`
	Value   string   `xml:",chardata"`
}
type resultFieldProps struct {
	XMLName     xml.Name `xml:"Field"`
	Name        string   `xml:"name,omitempty,attr"`
	DataTypeOID uint32   `xml:"dtId,omitempty,attr"`
}
type postgresQuery struct {
	XMLName           xml.Name `xml:"Query"`
	ConvertNull2Empty bool     `xml:"convertNull2Empty,omitempty,attr"`
	GenerateHTML      bool     `xml:"generateHTML,omitempty,attr"`
	Connection        string   `xml:"connection,omitempty,attr"`
	Timeout           int64    `xml:"timeout,omitempty,attr"`
	Error             string   `xml:"error,omitempty,attr"`
	ExecutionTime     int64    `xml:"execTime,omitempty,attr"`
	FetchTime         int64    `xml:"fetchTime,omitempty,attr"`
	GSTime            int64    `xml:"gsTime,omitempty,attr"`

	RowsCount int `xml:"rowsCount,omitempty,attr"`
	SQL       struct {
		Text string `xml:",cdata"`
	} `xml:"SQL"`
	Params struct {
		Param []queryParam
	} `xml:"Params"`

	Result struct {
		Fields struct {
			Field []resultFieldProps
		} `xml:"Fields"`
		Rows struct {
			Count int `xml:"count,attr"`
			Row   []dataRow
		} `xml:"Rows"`
		HTML struct {
			Text string `xml:",cdata"`
		} `xml:"HTML"`
	} `xml:"Result"`
}

const trackerConnectionName = "tracker"

func doPostgresQuery(w http.ResponseWriter, r *http.Request) {
	var q postgresQuery
	err := xml.NewDecoder(r.Body).Decode(&q)
	if err == nil {
		query2xml(&q)
	} else {
		q.Error = "Error parsing input XML"
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	//copier.CopyWithOption(&q1, &q, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	xml.NewEncoder(w).Encode(&q)
	return
}

func query2html(q *postgresQuery) string {
	qr := *q
	txtQuery := textHTML("<span style='color:black;font-weight:bold;'>###</span>")
	txtValue := textHTML("<span style='color:blue;font-weight:bold1;'>###</span>")
	txtError := textHTML("<span style='color:red;font-weight:bold1;'>###</span>")
	txtParamName := textHTML("<span style='color:black;font-weight:bold1;'>###</span>")
	txtParamValue := textHTML("<span style='color:blue;'>###</span>")

	td := tdHTML()
	tr := trHTML()
	rez := txtQuery(qr.SQL.Text)

	for _, p := range qr.Params.Param {
		rez += "<br>" + txtParamName(p.Name) + "{" + p.Type + "}=" + txtParamValue(p.Value)
	}
	rez += "<p>"
	rez += txtParamName("Exec Time") + ": " + txtValue(qr.ExecutionTime)
	rez += ", " + txtParamName("Fetch Time") + ": " + txtValue(qr.FetchTime)
	rez += ", " + txtParamName("Rows") + ": " + txtValue(qr.RowsCount)
	rez += ", " + txtParamName("Error") + ": " + txtError(qr.Error)
	rez += "<p>"
	rez += htmlTableEnd()
	headers := make([]interface{}, 0)
	for _, f := range qr.Result.Fields.Field {
		headers = append(headers, f.Name)
	}

	rez += "</p>"
	rez += htmlTableStart(headers...)
	for _, r := range qr.Result.Rows.Row {
		s := ""
		for _, f := range r.Field {
			s += td(strings.ReplaceAll(f.Value, "\x0b", "\r"))
		}
		s = tr(s)
		rez += s
	}
	rez += htmlTableEnd()
	return rez
}

func getConnection(name string) (*pgxpool.Pool, error) {
	conn := mapOfDBPools[name]
	if conn == nil {
		err := fmt.Errorf("DB connection not found!")
		return nil, err
	}
	return conn, nil
}

func query2xml(q *postgresQuery) error {
	(*q).Error = ""
	args := pgx.NamedArgs{}
	for _, arg := range (*q).Params.Param {
		v, err := buildBinaryValueFromStringValue(arg.Value, arg.Type, arg.Format)
		if err != nil {
			(*q).Error = err.Error()
			return err
		}
		args[arg.Name] = v
	}
	conn := mapOfDBPools[(*q).Connection]
	if conn == nil {
		err := fmt.Errorf("DB connection not found!")
		(*q).Error = err.Error()
		return err
	}
	startTime := time.Now()
	ctxBackground := context.Background()
	timeout := (*q).Timeout
	if timeout == 0 {
		timeout = 300
	}
	ctx, cancel := context.WithTimeout(ctxBackground, time.Duration(timeout)*time.Second)
	defer cancel()
	rows, err := conn.Query(ctx, (*q).SQL.Text, args)
	doneTime := time.Now()
	(*q).ExecutionTime = (doneTime.Sub(startTime)).Milliseconds()
	startTime = time.Now()
	if err != nil {
		(*q).Error = err.Error()
		return err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	numFields := len(fieldDescriptions)

	for i := 0; i < numFields; i++ {
		var fieldProps resultFieldProps
		fieldProps.Name = fieldDescriptions[i].Name
		fieldProps.DataTypeOID = fieldDescriptions[i].DataTypeOID
		(*q).Result.Fields.Field = append((*q).Result.Fields.Field, fieldProps)
	}
	idx := 0
	for rows.Next() {
		var dr dataRow
		idx++
		dr.Index = idx
		rowValues, err := rows.Values()
		if err != nil {
			(*q).Error = err.Error()
			return err
		}
		var f dataField
		for i := 0; i < numFields; i++ {
			f.Value = postgresValue2Str(rowValues[i], fieldDescriptions[i].DataTypeOID, (*q).ConvertNull2Empty)
			f.Name = fieldDescriptions[i].Name
			dr.Field = append(dr.Field, f)
		}
		(*q).Result.Rows.Row = append((*q).Result.Rows.Row, dr)
	}
	(*q).Result.Rows.Count = idx
	(*q).RowsCount = idx
	doneTime = time.Now()
	(*q).FetchTime = (doneTime.Sub(startTime)).Milliseconds()
	if err := rows.Err(); err != nil {
		(*q).Error = err.Error()
		return err
	}
	return nil
}

func checkDBConnection(name, query string) string {
	var q postgresQuery
	q.Connection = name
	q.SQL.Text = query
	if query == "" {
		return ""
	}
	query2xml(&q)
	if q.Error != "" {
		return q.Error
	}
	return ""
}

func checkDBConnections() string {
	for _, c := range svcConfig.Connections.Connection {
		if c.TestQuery == "" {
			continue
		}
		rez := checkDBConnection(c.Name, c.TestQuery)
		if rez != "" {
			return "ERROR: " + rez
		}
	}
	return "OK"
}

func updateToken(token *Token, expectedStatus []int, newStatus int, action, comment string) error {
	isStatusOK := func(status int) bool {
		for _, st := range expectedStatus {
			if status == st {
				return true
			}
		}
		return false
	}

	id := token.ID

	conn, err := getConnection(trackerConnectionName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	query := `select status from token where id=@id`
	args := pgx.NamedArgs{
		"id": id,
	}

	var status int

	err = conn.QueryRow(ctx, query, args).Scan(&status)
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	if !isStatusOK(status) {
		return fmt.Errorf("token update error. token status: %v ", status)
	}

	// =================
	tx, err := conn.Begin(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer tx.Rollback(ctx)
	status = 2
	query = `UPDATE token set status=@status, xml=@xml where id=@id`
	args = pgx.NamedArgs{
		"id":     id,
		"status": newStatus,
		"xml":    XML2String(*token),
	}

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("unable to execute query: %w", err)
	}

	query = `INSERT INTO log (id, action, status, comment, xml) VALUES(@id,@action, @status,@comment, @xml)`
	args = pgx.NamedArgs{
		"id":      id,
		"action":  action,
		"status":  newStatus,
		"comment": comment,
		"xml":     XML2String(*token),
	}

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("unable to save log: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//=================

	return nil
}

func createToken(token *Token, comment string) error {
	conn, err := getConnection(trackerConnectionName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	// =================
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	id := token.ID
	status := 1 //new
	query := `insert into token (id, status, xml) values (@id, @status, @xml)`
	args := pgx.NamedArgs{
		"id":     id,
		"status": status,
		"xml":    XML2String(*token),
	}

	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("unable to execute query: %w", err)
	}

	query = `INSERT INTO log (id, action, status, comment, xml) VALUES(@id,@action, @status,@comment, @xml)`
	args = pgx.NamedArgs{
		"id":      id,
		"action":  "create",
		"status":  status,
		"comment": comment,
		"xml":     XML2String(*token),
	}
	_, err = tx.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("unable to save log: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//=================

	return nil
}

func getToken(id string) (*Token, error) {
	query := `select status, xml from token where id=@id`
	args := pgx.NamedArgs{
		"id": id,
	}
	conn, err := getConnection(trackerConnectionName)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	var xmlText string
	var status int

	err = conn.QueryRow(ctx, query, args).Scan(&status, &xmlText)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	var token Token
	err = xml.Unmarshal([]byte(xmlText), &token)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	token.Status = strconv.Itoa(status)
	token.EditMessage = token.Message1 //can edit msg
	if token.Status == "3" {           //locked
		token.Locked = true
		token.EditMessage = token.Message3 //locked msg
	}

	return &token, nil
}
