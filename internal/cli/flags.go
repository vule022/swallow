package cli

import "github.com/spf13/cobra"

const (
	FlagModel       = "model"
	FlagDetailLevel = "detail-level"
)

// AddModelFlag adds the --model flag to a command.
func AddModelFlag(cmd *cobra.Command) {
	cmd.Flags().String(FlagModel, "", "override the LLM model (e.g. gpt-4o, claude-3-5-sonnet-20241022)")
}

// AddDetailLevelFlag adds the --detail-level flag to a command.
func AddDetailLevelFlag(cmd *cobra.Command) {
	cmd.Flags().String(FlagDetailLevel, "", "plan detail level: compact, standard, or detailed")
}

// GetModelOverride returns the --model flag value.
func GetModelOverride(cmd *cobra.Command) string {
	v, _ := cmd.Flags().GetString(FlagModel)
	return v
}

// GetDetailLevelStr returns the raw --detail-level string value.
func GetDetailLevelStr(cmd *cobra.Command) string {
	v, _ := cmd.Flags().GetString(FlagDetailLevel)
	return v
}
