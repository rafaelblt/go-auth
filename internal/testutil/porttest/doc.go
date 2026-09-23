// Package porttest holds one hand-written fake per port. Readers are in-memory
// stores a test seeds, writers record what they were given, SetError makes a
// fake fail, and FakeClock fixes the time.
//
// See docs/development/testing.md#fakes.
package porttest
