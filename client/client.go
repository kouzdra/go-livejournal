// forked from github.com/mattn/go-xmlrpc
package client

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"io"
	"encoding/hex"
	"crypto/md5"
)

type Array []interface{}
type Struct map[string]interface{}

type Client struct {
	//client *http.Client
	Url      string
	UserName string
	Password string
}

var xmlSpecial = map[byte]string{
	'<':  "&lt;",
	'>':  "&gt;",
	'"':  "&quot;",
	'\'': "&apos;",
	'&':  "&amp;",
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		if s, ok := xmlSpecial[c]; ok {
			b.WriteString(s)
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

type valueNode struct {
	Type string `xml:"attr"`
	Body string `xml:"chardata"`
}

func next(p *xml.Decoder) (interface{}, error) {
	se := nextStart(p)
	var nv interface{}
	switch se.Name.Local {
	case "string":
		var s string
		if e := p.DecodeElement(&s, &se); e != nil {
			return nil, e
		}
		return s, nil
	case "boolean":
		var s string
		if e := p.DecodeElement(&s, &se); e != nil {
			return nil, e
		}
		s = strings.TrimSpace(s)
		var b bool
		switch s {
		case "true","1":
			b = true
		case "false","0":
			b = false
		default:
			panic (errors.New("invalid boolean value"))
		}
		return b, nil
	case "int", "i1", "i2", "i4", "i8":
		var s string
		var i int
		if e := p.DecodeElement(&s, &se); e != nil {
			return nil, e
		}
		i, e := strconv.Atoi(strings.TrimSpace(s))
		return i, e
	case "double":
		var s string
		var f float64
		if e := p.DecodeElement(&s, &se); e != nil {
			return nil, e
		}
		f, e := strconv.ParseFloat(strings.TrimSpace(s), 64)
		return f, e
	case "dateTime.iso8601":
		var s string
		if e := p.DecodeElement(&s, &se); e != nil {
			return nil, e
		}
		t, e := time.Parse("20060102T15:04:05", s)
		if e != nil {
			t, e = time.Parse("2006-01-02T15:04:05-07:00", s)
			if e != nil {
				t, e = time.Parse("2006-01-02T15:04:05", s)
			}
		}
		return t, e
	case "base64":
		var s string
		if e := p.DecodeElement(&s, &se); e != nil {
			return nil, e
		}
		if b, e := base64.StdEncoding.DecodeString(s); e != nil {
			return nil, e
		} else {
			return b, nil
		}
	case "member":
		nextStart(p)
		return next(p)
	case "value":
		nextStart(p)
		return next(p)
	case "name":
		nextStart(p)
		return next(p)
	case "struct":
		st := Struct{}

		se = nextStart(p)
		for se.Name.Local == "member" {
			// name
			se = nextStart(p)
			if se.Name.Local != "name" {
				return nil, errors.New("invalid response")
			}
			var name string
			if e := p.DecodeElement(&name, &se); e != nil {
				panic (e)
			}
			se = nextStart(p)
			value, e := next(p)
			if se.Name.Local != "value" {
				return nil, errors.New("invalid response")
			}
			if e != nil {
				break
			}
			st [name] = value
			se = nextStart (p)
		}
		return st, nil
	case "array":
		var ar Array
		nextStart(p) // data
		for {
			nextStart(p) // top of value
			value, e := next(p)
			if e != nil {
				break
			}
			ar = append(ar, value)
		}
		return ar, nil
	}

	if e := p.DecodeElement(nv, &se); e != nil {
		return nil, e
	}
	return nv, nil
}
func nextStart(p *xml.Decoder) xml.StartElement {
	for {
		t, e := p.Token()
		if e != nil {
			if fmt.Sprintf ("%s", e) == "EOF" {
				return xml.StartElement{}
			}
			panic (e)
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t
		}
	}
	panic("unreachable")
}

func to_xml(v interface{}, typ bool) (s string) {
	r := reflect.ValueOf(v)
	t := r.Type()
	k := t.Kind()

	if b, ok := v.([]byte); ok {
		return "<base64>" + base64.StdEncoding.EncodeToString(b) + "</base64>"
	}

	switch k {
	case reflect.Invalid:
		panic("unsupported type")
	case reflect.Bool:
		return fmt.Sprintf("<boolean>%v</boolean>", v)
	case reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if typ {
			return fmt.Sprintf("<int>%v</int>", v)
		}
		return fmt.Sprintf("%v", v)
	case reflect.Uintptr:
		panic("unsupported type")
	case reflect.Float32, reflect.Float64:
		if typ {
			return fmt.Sprintf("<double>%v</double>", v)
		}
		return fmt.Sprintf("%v", v)
	case reflect.Complex64, reflect.Complex128:
		panic("unsupported type")
	case reflect.Array:
		s = "<array><data>"
		for n := 0; n < r.Len(); n++ {
			s += "<value>"
			s += to_xml(r.Index(n).Interface(), typ)
			s += "</value>"
		}
		s += "</data></array>"
		return s
	case reflect.Chan:
		panic("unsupported type")
	case reflect.Func:
		panic("unsupported type")
	case reflect.Interface:
		return to_xml(r.Elem(), typ)
	case reflect.Map:
		s = "<struct>"
		for _, key := range r.MapKeys() {
			s += "<member>"
			s += "<name>" + xmlEscape(key.Interface().(string)) + "</name>"
			s += "<value>" + to_xml(r.MapIndex(key).Interface(), typ) + "</value>"
			s += "</member>"
		}
		s += "</struct>"
		return s
	case reflect.Ptr:
		panic("unsupported type")
	case reflect.Slice:
		panic("unsupported type")
	case reflect.String:
		if typ {
			return fmt.Sprintf("<string>%v</string>", xmlEscape(v.(string)))
		}
		return xmlEscape(v.(string))
	case reflect.Struct:
		s = "<struct>"
		for n := 0; n < r.NumField(); n++ {
			s += "<member>"
			s += "<name>" + (string)(t.Field(n).Tag) + "</name>"
			s += "<value>" + to_xml(r.FieldByIndex([]int{n}).Interface(), true) + "</value>"
			s += "</member>"
		}
		s += "</struct>"
		return s
	case reflect.UnsafePointer:
		return to_xml(r.Elem(), typ)
	}
	return
}

type Auth struct {
	Username  string "username"
	Password  string "password"
	Hpassword string "hpassword"
	Ver          int "ver"
}

func (client *Client) Call(name string, args ... interface {}) Struct {
	s := "<methodCall>"
	s += "<methodName>" + xmlEscape("LJ.XMLRPC." + name) + "</methodName>"
	s += "<params>"

	addArg := func (arg interface{}) {
		s += "<param><value>"
		s += to_xml(arg, true)
		s += "</value></param>"
	}

	hash := md5.New ()
	io.WriteString (hash, client.Password)
	addArg (Auth {
		Username: client.UserName,
		//Password: client.Password,
		//Hpassword: "",
		Password: "",
		Hpassword: hex.EncodeToString (hash.Sum (nil)),
		Ver: 1,
	})
	for arg := range args {
		addArg (arg)
	}
	s += "</params></methodCall>"
	fmt.Printf ("Req:\n%s\n", s);
	bs := bytes.NewBuffer([]byte(s))
	r, e := http.Post(client.Url, "text/xml", bs)
	if e != nil {
		panic (&HTTPError {Error:e})
	}
	defer r.Body.Close()

	p := xml.NewDecoder(r.Body)
	nextReq := func (req string) {
		se := nextStart(p) // methodResponse
		if se.Name.Local != req {
			panic (&Format {Req: req, Token: se.Name.Local})
		}
	}

	nextReq ("methodResponse");
	se := nextStart(p) // params
	if se.Name.Local == "params" {
		nextReq ("param");
		nextReq ("value");
		v, e := next(p)
		if e != nil {
			panic (e)
		}
		return v.(Struct)
	} else if se.Name.Local == "fault" {
		nextReq ("value");
		v, e := next(p)
		if e != nil {
			panic (e)
		}
		s := v.(Struct)
		m := s ["faultString"].(string)
		c := s ["faultCode"].(int)
		panic (&Fault {Code: c, Message: m})
	} else {
		panic (&Format {Token: se.Name.Local, Req: "params\" or \"fault"})
	}
}
