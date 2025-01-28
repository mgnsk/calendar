package snowflake

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// Generate a new snowflake ID.
func Generate() snowflake.ID {
	return snowflakeNode.Generate()
}

var snowflakeNode *snowflake.Node

func init() {
	// Define a custom epoch, the start time of Raskemuusikaliit's general meeting
	// where the calendar idea was conceived.
	epoch, err := time.Parse(time.RFC3339, "2025-01-27T18:30:00+02:00") // TODO: offset correct?
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
