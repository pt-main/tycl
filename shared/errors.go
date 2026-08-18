package shared

import (
	"github.com/pt-main/lc/public/errors"
)

const (
	ContextedError  errors.ErrorCodeType = "CONTEXTED"
	RuntimeError    errors.ErrorCodeType = "RUNTIME_ERROR"
	ProcessingError errors.ErrorCodeType = "PROCESSING_ERROR"
	WrappedError    errors.ErrorCodeType = "WRAPPED"
)
