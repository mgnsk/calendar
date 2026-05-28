package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// ID is the snowflake ID.
type ID snowflake.ID

func (id ID) String() string {
	return snowflake.ID(id).String()
}

// Int64 returns an int64 of the snowflake ID
func (id ID) Int64() int64 {
	return snowflake.ID(id).Int64()
}

// UnmarshalText unmarshals the ID from text.
func (id *ID) UnmarshalText(text []byte) error {
	v, err := snowflake.ParseBytes(text)
	if err != nil {
		return err
	}

	*id = ID(v)

	return nil
}

// Generate a new snowflake ID.
func Generate() ID {
	return ID(snowflakeNode.Generate())
}

// ParseTime parses time (in local time) from snowflake ID.
func ParseTime(id int64) time.Time {
	unixMillis := snowflake.ParseInt64(id).Time()

	return time.UnixMilli(unixMillis)
}

var snowflakeNode *snowflake.Node

func init() {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	snowflakeNode = node
}
