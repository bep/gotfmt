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
	_ = x[tBracketOpen-3]
	_ = x[tBracketClose-4]
	_ = x[tSpace-5]
	_ = x[tNewline-6]
	_ = x[tOther-7]
	_ = x[tAction-8]
	_ = x[tComment-9]
	_ = x[tActionStart-10]
	_ = x[tActionEndStart-11]
	_ = x[tActionEnd-12]
}

const _itemType_name = "tZerotErrortEOFtBracketOpentBracketClosetSpacetNewlinetOthertActiontCommenttActionStarttActionEndStarttActionEnd"

var _itemType_index = [...]uint8{0, 5, 11, 15, 27, 40, 46, 54, 60, 67, 75, 87, 102, 112}

func (i itemType) String() string {
	if i < 0 || i >= itemType(len(_itemType_index)-1) {
		return "itemType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}
