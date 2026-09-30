package validation

import "regexp"

// ItemNameRegex is the shared rule for item names (warehouse items, store stock
// search, category item search and order item snapshots). Keep them in sync by
// always using this variable.
//
//   - Must start with a letter or digit (any language)
//   - Followed by letters, digits, spaces, punctuation (& - . , ' ( ) / % # ...),
//     currency symbols ($ ¥ €), other symbols (© ® ™ °) and '+'
//   - Math-like symbols such as < > = | ~ ^ ` are not allowed
//   - Max 255 characters
var ItemNameRegex = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N}\p{Zs}\p{P}\p{Sc}\p{So}+]{0,254}$`)
