package parser

import (
	"github.com/rafecolton/docker-builder/parser/uuid"
)

/*
NextUUID returns the next UUID generated by the parser's uuid generator.  This
will either be a random uuid (normal behavior) or the same uuid every time if
the generator is "seeded" (used for tests)
*/
func (parser *Parser) NextUUID() (string, error) {
	return parser.uuidGenerator.NextUUID()
}

/*
SeedUUIDGenerator turns this parser's uuidGenerator into a seeded generator.
All calls to NextUUID() will produce the same uuid after this function is
called and until RandomizeUUIDGenerator() is called.
*/
func (parser *Parser) SeedUUIDGenerator() {
	parser.uuidGenerator = uuid.NewUUIDGenerator(false)
}

/*
RandomizeUUIDGenerator turns this parser's uuidGenerator into a random
generator.  All calls to NextUUID() will produce a random uuid after this
function is called and until SeedUUIDGenerator() is called.
*/
func (parser *Parser) RandomizeUUIDGenerator() {
	parser.uuidGenerator = uuid.NewUUIDGenerator(true)
}