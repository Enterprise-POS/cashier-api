package validation

import "regexp"

// UserNameRegex is the shared rule for user names and responsible-user name
// snapshots. Keep registration and transaction validation aligned.
var UserNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_ ]{2,255}$`)
