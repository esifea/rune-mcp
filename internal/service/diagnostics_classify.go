package service

import (
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RunespaceErrorType — runespace probe error classification.
// Python: server.py:L655-672 (string pattern matching — Python).
// Go: gRPC status.Code() enum based.
type RunespaceErrorType string

const (
	RunespaceErrConnectionRefused RunespaceErrorType = "connection_refused"
	RunespaceErrAuthFailure       RunespaceErrorType = "auth_failure"
	RunespaceErrDeadlineExceeded  RunespaceErrorType = "deadline_exceeded"
	RunespaceErrTimeout           RunespaceErrorType = "timeout" // context.WithTimeout deadline
	RunespaceErrUnknown           RunespaceErrorType = "unknown"
)

// ClassifyRunespaceError maps an error (with its elapsed latency) to a typed
// classification + user-facing hint. Used by LifecycleService.Diagnostics.
//
// Python match (server.py:L655-672):
//
//	UNAVAILABLE | Connection refused → connection_refused
//	UNAUTHENTICATED | 401             → auth_failure
//	DEADLINE_EXCEEDED                  → deadline_exceeded
//	other                              → unknown
//
// Hints (Python exact strings, keep bit-identical for schema)
// XXX: it seems that ErrDeadlineExcceded can cover ErrTimeout
func ClassifyRunespaceError(err error, elapsed time.Duration) (RunespaceErrorType, string) {
	st, ok := status.FromError(err)
	if !ok {
		return RunespaceErrUnknown, fmt.Sprintf("Unexpected runespace error (%.1fms): %v", float64(elapsed.Milliseconds()), err)
	}

	switch st.Code() {
	case codes.Unavailable:
		return RunespaceErrConnectionRefused, "Runespace cluster appears unreachable from this host - check network connectivity"
	case codes.Unauthenticated:
		return RunespaceErrAuthFailure, "Runespace API key was rejected - contact your Vault administrator"
	case codes.DeadlineExceeded:
		return RunespaceErrDeadlineExceeded, "Runespace gRPC deadline exceeded - check network latency to the cluster"
	default:
		return RunespaceErrUnknown, "Runespace probe failed after recovery attempt - check network connectivity"
	}
}
