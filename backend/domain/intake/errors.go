package intake

import "errors"

// ErrInvestigateDisabled is returned when investigate is disabled by server configuration.
var ErrInvestigateDisabled = errors.New("investigate disabled")
