package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// ID is the snowflake ID.
type ID = snowflake.ID

// Generate a new snowflake ID.
func Generate() ID {
	return snowflakeNode.Generate()
}

// ParseTime parses time (in local time) from snowflake ID.
func ParseTime(id int64) time.Time {
	unixMillis := snowflake.ParseInt64(id).Time()

	return time.UnixMilli(unixMillis)
}

var snowflakeNode *snowflake.Node

func init() {
	// Define a custom epoch, the start time of Raskemuusikaliit's general meeting
	// where the calendar idea was conceived.
	epoch, err := time.Parse(time.RFC3339, "2025-01-27T18:30:00+02:00")
	if err != nil {
		panic(err)
	}

	snowflake.Epoch = epoch.UnixMilli()

	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	snowflakeNode = node
}
