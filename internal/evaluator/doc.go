// Package evaluator decodes, validates, and grades the frozen public Level 7
// semantic-evaluation protocol. Production code is pure: callers supply every
// byte and observed fact, and the package exposes no filesystem, process,
// network, clock, randomness, environment, credential, logging, or mutation
// interface.
package evaluator
