package gbs

import "errors"

var (
	ErrXMLDecode = errors.New("xml decode error")
	ErrDatabase  = errors.New("database error")
)
