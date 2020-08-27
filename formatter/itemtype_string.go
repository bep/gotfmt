// Code generated by "stringer -type itemType"; DO NOT EDIT.

package formatter

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[tZero-0]
	_ = x[tError-1]
	_ = x[tEOF-2]
	_ = x[tAction-3]
	_ = x[tComment-4]
	_ = x[tActionStart-5]
	_ = x[tActionEndStart-6]
	_ = x[tActionEnd-7]
	_ = x[tBracketOpen-8]
	_ = x[tBracketClose-9]
	_ = x[tSpace-10]
	_ = x[tQuoteStart-11]
	_ = x[tQuoteEnd-12]
	_ = x[tNewline-13]
	_ = x[tOther-14]
}

const _itemType_name = "tZerotErrortEOFtActiontCommenttActionStarttActionEndStarttActionEndtBracketOpentBracketClosetSpacetQuoteStarttQuoteEndtNewlinetOther"

var _itemType_index = [...]uint8{0, 5, 11, 15, 22, 30, 42, 57, 67, 79, 92, 98, 109, 118, 126, 132}

func (i itemType) String() string {
	if i < 0 || i >= itemType(len(_itemType_index)-1) {
		return "itemType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}
