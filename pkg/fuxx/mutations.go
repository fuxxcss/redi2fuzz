package fuxx

import (
	"strings"
)

// interest
const (
	InterestEmpty int = iota
	InterestNULL
	InterestTerminal
	InterestHex
	InterestSpecial
	InterestLong
)

// interesting strings
var	InterestingStr = []string {
	"", 					 // empty
	"\x00",					 // null
	"\r",					 // terminal
	"\xff\xfe",				 // hex
	"\";+-*>([",			 // special
	strings.Repeat("a", 100000), // long str
}

// interesting nums
var InterestingNum = []string {
	"-128",
	"-1",
	"-0",
	"0",   
	"1",   
	"127",
	"255",
	"-32768",
	"32767",
	"65535", 
	"-2147483648",
	"2147483647",
	"9223372036854775807",
	"-9223372036854775808",
	"-0.0",
	"0.0",
	"-0.0000000000000001",
	"0.0000000000000001",
	"-3.14",
	"3.14",
	"-inf",
	"+inf",
	"nan",
}



