package dolores

import "errors"

var (
	ErrInvalidKeyFile  = errors.New("invalid key file")
	ErrInvalidFormat   = errors.New("invalid file content format")
	ErrInvalidIdentity = errors.New("invalid identity")
)
