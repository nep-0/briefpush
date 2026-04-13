package ai

import "context"

type AI interface {
	GenerateSummary(ctx context.Context, title string, content string) (string, error)
	GenerateReport(ctx context.Context, summaries []string) (string, error)
}
