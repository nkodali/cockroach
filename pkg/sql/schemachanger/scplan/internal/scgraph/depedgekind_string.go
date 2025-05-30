// Copyright 2023 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

// Code generated by "stringer"; DO NOT EDIT.

package scgraph

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Precedence-1]
	_ = x[SameStagePrecedence-2]
	_ = x[PreviousStagePrecedence-3]
	_ = x[PreviousTransactionPrecedence-4]
}

func (i DepEdgeKind) String() string {
	switch i {
	case Precedence:
		return "Precedence"
	case SameStagePrecedence:
		return "SameStagePrecedence"
	case PreviousStagePrecedence:
		return "PreviousStagePrecedence"
	case PreviousTransactionPrecedence:
		return "PreviousTransactionPrecedence"
	default:
		return "DepEdgeKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
