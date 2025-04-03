package utils

// fuxx modes
const (
	ModeDumb string = "dumb"
	ModeGramfree string = "gramfree"
	ModeFagent string = "fagent"
)

var Modes = map[string]interface{} {
	ModeDumb : fuxx.MutateDumb,
	ModeGramfree : fuxx.MutateGramfree,
	ModeFagent : fuxx.MutateFagent,
}


