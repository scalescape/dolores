package cerr

import "errors"

var (
	ErrNoProjectFound       = errors.New("no project found")
	ErrInvalidOrg           = errors.New("invalid org details")
	ErrInvalidEnvironment   = errors.New("invalid environment")
	ErrInvalidOrgID         = errors.New("invalid org id")
	ErrInvalidSecretRequest = errors.New("invalid secrets request")
	ErrNoSecretFound        = errors.New("no secrets found")
)
