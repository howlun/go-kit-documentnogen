package common

var (
	DefaultSeqNoFormat    = `%0*d` // leading * (variable) number of 0 (zero)
	DefaultSeqNoLength    = 5
	DefaultDocFormat      = "{{PREFIX}}{{DOCTYPE}}{{BRHCD}}{{YEAR}}{{SEQNO}}" // NOTE: the variable name should match this pattern: regexp.MustCompile(`{{[a-zA-Z]+}}`)
	MustCompilePatternStr = `{{[a-zA-Z]+}}`
	FixedVarPrefix        = "PREFIX"
	FixedVarSeqNo         = "SEQNO"
)
