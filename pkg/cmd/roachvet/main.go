// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

// Command roachvet is a vettool which includes all of the standard analysis
// passes included in go vet as well as the `shadow` pass and some first-party
// passes.
package main

import (
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/deferloop"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/deferunlockcheck"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/errcmp"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/errwrap"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/fmtsafe"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/forbiddenmethod"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/hash"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/leaktestcall"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/nilness"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/nocopy"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/redactcheck"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/returnerrcheck"
	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/unconvert"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	var as []*analysis.Analyzer
	// First-party analyzers:
	as = append(as, forbiddenmethod.Analyzers...)
	as = append(as,
		hash.Analyzer,
		leaktestcall.Analyzer,
		nocopy.Analyzer,
		redactcheck.Analyzer,
		returnerrcheck.Analyzer,
		unconvert.Analyzer,
		fmtsafe.Analyzer,
		errcmp.Analyzer,
		nilness.Analyzer,
		errwrap.Analyzer,
		deferunlockcheck.Analyzer,
		deferloop.Analyzer,
	)

	// Standard go vet analyzers:
	as = append(as,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	// Additional analyzers:
	as = append(as,
		shadow.Analyzer,
	)

	unitchecker.Main(as...)
}
