package fuxx

type InterestingStrType int

const (
	InterestEmpty InterestingStrType = iota
	InterestNULL
	InterestTerminal
	InterestHex
	InterestSpecial
	InterestShort
)

// interesting strings
var	InterestingStr = []string {
	"", 			// empty
	"\x00",			// null
	"\r",			// terminal
	"\xff\xfe",		// hex
	"\";+-*>([",	// special
	"a"*256, 		// short str
}

// interesting ints
var InterestingInt = []string {
	"-128",
	"-1",
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
}



