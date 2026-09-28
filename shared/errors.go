package shared

import (
	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/public/errors"
)

const (
	ContextedError  errors.ErrorCodeType = "CONTEXTED"
	RuntimeError    errors.ErrorCodeType = "RUNTIME_ERROR"
	ProcessingError errors.ErrorCodeType = "PROCESSING_ERROR"
	WrappedError    errors.ErrorCodeType = "WRAPPED"
	ContractError   errors.ErrorCodeType = "CONTRACT"
)

// HintError attaches an actionable suggestion to an error.
func HintError(err core.ErrorInterface, hint string) core.ErrorInterface {
	if ce, ok := err.(*core.Error); ok {
		ce.WithMeta("hint", hint)
	}
	return err
}

// StartError attaches a token range to an error, in file coordinates.
func StartError(err core.ErrorInterface, start, end int) core.ErrorInterface {
	if ce, ok := err.(*core.Error); ok {
		ce.WithMeta("start", start)
		ce.WithMeta("end", end)
	}
	return err
}
