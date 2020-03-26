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
	_ = x[Backslash-10]
	_ = x[Nullable-11]
	_ = x[OpenParen-12]
	_ = x[CloseParen-13]
	_ = x[OpenBrack-14]
	_ = x[CloseBrack-15]
	_ = x[OpenAngle-16]
	_ = x[CloseAngle-17]
	_ = x[Comma-18]
	_ = x[Ellipsis-19]
	_ = x[Union-20]
	_ = x[Intersect-21]
	_ = x[Ident-22]
}

const _TokenType_name = "EOFErrorNewlineWhitespaceAsteriskOtherOpenDocCloseDocTagVarBackslashNullableOpenParenCloseParenOpenBrackCloseBrackOpenAngleCloseAngleCommaEllipsisUnionIntersectIdent"

var _TokenType_index = [...]uint8{0, 3, 8, 15, 25, 33, 38, 45, 53, 56, 59, 68, 76, 85, 95, 104, 114, 123, 133, 138, 146, 151, 160, 165}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
