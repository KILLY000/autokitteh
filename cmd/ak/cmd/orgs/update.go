package orgs

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var updateCmd = common.StandardCommand(&cobra.Command{
	Use:   "update <org-id or name> [--display-name display-name] [--name name]",
	Short: "Update an org",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}
		oid, err := r.Org(ctx, args[0])
		if err != nil {
			return fmt.Errorf("resolve org: %w", err)
		}

		o := sdktypes.NewOrg().WithID(oid)

		fm := &sdktypes.FieldMask{}

		if cmd.Flags().Changed("display-name") {
			fm.Paths = append(fm.Paths, "display_name")
			o = o.WithDisplayName(displayName)
		}

		if cmd.Flags().Changed("name") {
			name, err := sdktypes.ParseSymbol(name)
			if err != nil {
				return fmt.Errorf("invalid name: %w", err)
			}

			fm.Paths = append(fm.Paths, "name")
			o = o.WithName(name)
		}

		if err := orgs().Update(ctx, o, fm); err != nil {
			return fmt.Errorf("update org: %w", err)
		}

		return nil
	},
})

func init() {
	updateCmd.Flags().StringVarP(&displayName, "display-name", "t", "", "org's display name")
	updateCmd.Flags().StringVarP(&name, "name", "n", "", "org's name")
}
