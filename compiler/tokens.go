// Copyright 2018 Alexey Krivonogov. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gentee

const (
	tkIdent  = iota + 1 // identifier
	tkLine              // a new line
	tkInt               // integer number (10-base)
	tkIntHex            // integer number (16-base)
	tkIntOct            // integer number (8-base)
	tkError             // tkError can be only the last tken
)

// Operators
const (
	tkAdd    = iota + 32 // +
	tkSub                // -
	tkMul                // *
	tkDiv                // /
	tkLPar               // (
	tkRPar               // )
	tkLCurly             // {
	tkRCurly             // }
	tkAssign             // =
)

// Keywords
const (
	tkRun     = iota + 64 // run
	tkDefault             // is used for preCompileTable
)
