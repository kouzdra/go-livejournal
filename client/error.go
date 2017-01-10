package client

import "fmt"

type Error interface {
	Show () string
}

type Fault struct {
	Code       int
	Message string
}

func (f Fault) Show () string {
	return fmt.Sprintf ("code #%d: %s", f.Code, f.Message)
}

type Format struct {
	Req   string
	Token string
}

func (f Format) Show () string {
	return fmt.Sprintf ("Response format: \"%s\", \"%s\" requred", f.Token, f.Req)
}

type HTTPError struct {
	Error error
}

func (f HTTPError) Show () string {
	return fmt.Sprintf ("%s", f.Error)
}

