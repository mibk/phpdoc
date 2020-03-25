// Code generated by "stringer -type TokenType"; DO NOT EDIT.

package phpdoc

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EOF-0]
	_ = x[Error-1]
	_ = x[Newline-2]
	_ = x[Whitespace-3]
	_ = x[Asterisk-4]
	_ = x[Other-5]
	_ = x[OpenDoc-6]
	_ = x[CloseDoc-7]
	_ = x[Tag-8]
	_ = x[Var-9]
	_ = x[OpenBrack-10]
	_ = x[CloseBrack-11]
	_ = x[Union-12]
	_ = x[Ident-13]
}

const _TokenType_name = "EOFErrorNewlineWhitespaceAsteriskOtherOpenDocCloseDocTagVarOpenBrackCloseBrackUnionIdent"

var _TokenType_index = [...]uint8{0, 3, 8, 15, 25, 33, 38, 45, 53, 56, 59, 68, 78, 83, 88}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
