// Copyright 2018 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package stdlib

import (
	"fmt"
	"strconv"

	"github.com/gentee/gentee/core"
)

// InitInt appends stdlib int functions to the virtual machine
func InitInt(ws *core.Workspace) {
	for _, item := range []interface{}{
		core.Link{floatºInt, 19<<16 | core.EMBED},        // float( int )
		core.Link{strºInt, 1<<16 | core.EMBED},           // str( int )
		core.Link{boolºInt, 3<<16 | core.EMBED},          // bool( int )
		core.Link{ExpStrºInt, 4<<16 | core.EMBED},        // expression in string
		core.Link{AssignºIntInt, core.ASSIGN},            // int = int
		core.Link{AssignºIntChar, core.ASSIGN},           // int = char
		core.Link{AssignAddºIntInt, core.ASSIGN + 1},     // int += int
		core.Link{AssignBitAndºIntInt, core.ASSIGN + 6},  // int &= int
		core.Link{AssignBitOrºIntInt, core.ASSIGN + 7},   // int |= int
		core.Link{AssignBitXorºIntInt, core.ASSIGN + 8},  // int ^= int
		core.Link{AssignDivºIntInt, core.ASSIGN + 4},     // int /= int
		core.Link{AssignModºIntInt, core.ASSIGN + 5},     // int %= int
		core.Link{AssignMulºIntInt, core.ASSIGN + 3},     // int *= int
		core.Link{AssignSubºIntInt, core.ASSIGN + 2},     // int -= int
		core.Link{AssignLShiftºIntInt, core.ASSIGN + 9},  // int <<= int
		core.Link{AssignRShiftºIntInt, core.ASSIGN + 10}, // int >>= int
		core.Link{MaxºIntInt, 37<<16 | core.EMBED},       // Max(int, int)
		core.Link{MinºIntInt, 38<<16 | core.EMBED},       // Min(int, int)
	} {
		ws.StdLib().NewEmbed(item)
	}
}

// AssignºIntInt assign one integer to another
func AssignºIntInt(ptr *interface{}, value int64) int64 {
	*ptr = value
	return (*ptr).(int64)
}

// AssignºIntChar assign a rune to integer
func AssignºIntChar(ptr *interface{}, value rune) int64 {
	*ptr = int64(value)
	return (*ptr).(int64)
}

// AssignAddºIntInt adds one integer to another
func AssignAddºIntInt(ptr *interface{}, value int64) (int64, error) {
	switch v := (*ptr).(type) {
	case uint8:
		value += int64(v)
		if uint64(value) > 255 {
			return 0, fmt.Errorf(core.ErrorText(core.ErrByteOut))
		}
		*ptr = value
	default:
		*ptr = v.(int64) + value
	}
	return (*ptr).(int64), nil
}

// AssignBitAndºIntInt equals int &= int
func AssignBitAndºIntInt(ptr *interface{}, value int64) int64 {
	*ptr = (*ptr).(int64) & value
	return (*ptr).(int64)
}

// AssignBitOrºIntInt equals int |= int
func AssignBitOrºIntInt(ptr *interface{}, value int64) int64 {
	*ptr = (*ptr).(int64) | value
	return (*ptr).(int64)
}

// AssignBitXorºIntInt equals int ^= int
func AssignBitXorºIntInt(ptr *interface{}, value int64) int64 {
	*ptr = (*ptr).(int64) ^ value
	return (*ptr).(int64)
}

// AssignDivºIntInt does int /= int
func AssignDivºIntInt(ptr *interface{}, value int64) (int64, error) {
	if value == 0 {
		return 0, fmt.Errorf(core.ErrorText(core.ErrDivZero))
	}
	*ptr = (*ptr).(int64) / value
	return (*ptr).(int64), nil
}

// AssignModºIntInt equals int %= int
func AssignModºIntInt(ptr *interface{}, value int64) (int64, error) {
	if value == 0 {
		return 0, fmt.Errorf(core.ErrorText(core.ErrDivZero))
	}
	*ptr = (*ptr).(int64) % value
	return (*ptr).(int64), nil
}

// AssignMulºIntInt equals int *= int
func AssignMulºIntInt(ptr *interface{}, value int64) int64 {
	*ptr = (*ptr).(int64) * value
	return (*ptr).(int64)
}

// AssignSubºIntInt equals int *= int
func AssignSubºIntInt(ptr *interface{}, value int64) int64 {
	*ptr = (*ptr).(int64) - value
	return (*ptr).(int64)
}

// AssignLShiftºIntInt does int <<= int
func AssignLShiftºIntInt(ptr *interface{}, value int64) (int64, error) {
	if value < 0 {
		return 0, fmt.Errorf(core.ErrorText(core.ErrShift))
	}
	*ptr = (*ptr).(int64) << uint64(value)
	return (*ptr).(int64), nil
}

// AssignRShiftºIntInt does int >>= int
func AssignRShiftºIntInt(ptr *interface{}, value int64) (int64, error) {
	if value < 0 {
		return 0, fmt.Errorf(core.ErrorText(core.ErrShift))
	}
	*ptr = (*ptr).(int64) >> uint64(value)
	return (*ptr).(int64), nil
}

// MaxºIntInt returns the maximum of two integers
func MaxºIntInt(left, right int64) int64 {
	if left < right {
		return right
	}
	return left
}

// MinºIntInt returns the minimum of two integers
func MinºIntInt(left, right int64) int64 {
	if left > right {
		return right
	}
	return left
}

// floatºInt converts integer value to float
func floatºInt(val int64) float64 {
	return float64(val)
}

// strºInt converts integer value to string
func strºInt(val int64) string {
	return strconv.FormatInt(val, 10)
}

// boolºInt converts integer value to boolean 0->false, not 0 -> true
func boolºInt(val int64) bool {
	return val != 0
}

// ExpStrºInt adds string and integer in string expression
func ExpStrºInt(left string, right int64) string {
	return left + strºInt(right)
}
