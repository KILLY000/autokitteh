package sessions

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// skip       int
var logOrder string

var logCmd = common.StandardCommand(&cobra.Command{
	Use:   "log [sessions ID | project] [--fail] [--skip <N>] [--no-timestamps]",
	Short: "Get session runtime logs (prints, calls, errors, state changes)",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		f := sdkservices.SessionLogRecordsFilter{SessionID: sid}
		if nextPageToken != "" {
			f.PageToken = nextPageToken
		}

		if pageSize > 0 {
			f.PageSize = int32(pageSize)
		}

		if skipRows > 0 {
			f.Skip = int32(skipRows)
		}

		f.Ascending = true
		if logOrder == "desc" {
			f.Ascending = false
		}

		return sessionLog(f)
	},
})

func init() {
	// Command-specific flags.
	// logCmd.Flags().IntVarP(&skip, "skip", "s", 0, "number of entries to skip")
	logCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from watch output")
	logCmd.Flags().StringVarP(&logOrder, "order", "o", "desc", "logs order can be asc or desc")
	logCmd.Flags().StringVar(&nextPageToken, "next-page-token", "", "provide the returned page token to get next")
	logCmd.Flags().IntVar(&pageSize, "page-size", 20, "page size")
	logCmd.Flags().IntVar(&skipRows, "skip-rows", 0, "skip rows")

	common.AddFailIfNotFoundFlag(logCmd)
}

// skip >= 0: skip first records
// skip < 0: skip all up to last |skip| records.
func sessionLog(filter sdkservices.SessionLogRecordsFilter) error {
	ctx, done := common.LimitedContext()
	defer done()

	l, err := sessions().GetLog(ctx, filter)
	if err != nil {
		return fmt.Errorf("get log: %w", err)
	}

	rs := l.Records
	if len(rs) == 0 {
		return nil
	}

	slices.SortFunc(rs, func(a, b sdktypes.SessionLogRecord) int {
		return a.Timestamp().Compare(b.Timestamp())
	})

	printLogs(rs)

	return nil
}

func printLogs(logs []sdktypes.SessionLogRecord) {
	for _, r := range logs {
		if noTimestamps {
			r = r.WithoutTimestamp().WithProcessID("")
		}

		if !quiet {
			common.Render(r)
		}
	}
}
