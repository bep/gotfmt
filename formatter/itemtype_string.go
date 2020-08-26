// Code generated by "stringer -type itemType"; DO NOT EDIT.

package formatter

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[tError-0]
	_ = x[tEOF-1]
	_ = x[tAction-2]
	_ = x[tComment-3]
	_ = x[tActionStart-4]
	_ = x[tActionEndStart-5]
	_ = x[tActionEnd-6]
	_ = x[tSpace-7]
	_ = x[tNewline-8]
	_ = x[tOther-9]
}

const _itemType_name = "tErrortEOFtActiontCommenttActionStarttActionEndStarttActionEndtSpacetNewlinetOther"

var _itemType_index = [...]uint8{0, 6, 10, 17, 25, 37, 52, 62, 68, 76, 82}

func (i itemType) String() string {
	if i < 0 || i >= itemType(len(_itemType_index)-1) {
		return "itemType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _itemType_name[_itemType_index[i]:_itemType_index[i+1]]
}
