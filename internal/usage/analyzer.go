package usage

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/Priyans-hu/argus/pkg/types"
)

// Options controls usage analysis behavior
type Options struct {
	Since            time.Time
	IncludeSubagents bool
	MaxWorkers       int // 0 = NumCPU
}

// AnalyzeUsage discovers and parses Claude Code session logs for a project,
// returning aggregated usage insights.
func AnalyzeUsage(ctx context.Context, projectPath string, opts Options) (*types.UsageInsights, error) {
	// Discover Claude project directory
	projectDir := DiscoverProjectDir(projectPath)
	if projectDir == "" {
		slog.Debug("no Claude Code session data found", "projectPath", projectPath)
		return nil, nil
	}

	// Discover session files
	sessionFiles, err := DiscoverSessions(projectDir, opts.Since)
	if err != nil {
		return nil, err
	}

	if len(sessionFiles) == 0 {
		slog.Debug("no session files found", "projectDir", projectDir)
		return nil, nil
	}

	slog.Debug("discovered sessions", "count", len(sessionFiles), "projectDir", projectDir)

	// Collect all files to parse (sessions + subagents)
	var allFiles []string
	allFiles = append(allFiles, sessionFiles...)

	if opts.IncludeSubagents {
		for _, sf := range sessionFiles {
			sessionID := sessionIDFromPath(sf)
			subagents, err := DiscoverSubagents(projectDir, sessionID)
			if err != nil {
				slog.Debug("error discovering subagents", "session", sessionID, "error", err)
				continue
			}
			allFiles = append(allFiles, subagents...)
		}
	}

	// Parse sessions concurrently
	workers := opts.MaxWorkers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > len(allFiles) {
		workers = len(allFiles)
	}

	parseOpts := ParseOptions{Since: opts.Since}
	sessions, err := parseSessionsConcurrently(ctx, allFiles, parseOpts, workers)
	if err != nil {
		return nil, err
	}

	// Aggregate
	insights := Aggregate(sessions, projectPath)

	return insights, nil
}

func parseSessionsConcurrently(ctx context.Context, files []string, opts ParseOptions, workers int) ([]*SessionData, error) {
	type result struct {
		session *SessionData
		err     error
	}

	fileCh := make(chan string, len(files))
	resultCh := make(chan result, len(files))

	// Feed files
	for _, f := range files {
		fileCh <- f
	}
	close(fileCh)

	// Spawn workers
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileCh {
				select {
				case <-ctx.Done():
					resultCh <- result{err: ctx.Err()}
					return
				default:
				}

				session, err := ParseSession(ctx, file, opts)
				resultCh <- result{session: session, err: err}
			}
		}()
	}

	// Close results when all workers done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var sessions []*SessionData
	for r := range resultCh {
		if r.err != nil {
			slog.Debug("error parsing session", "error", r.err)
			continue // skip failed sessions, don't abort
		}
		if r.session != nil && (r.session.TurnCount > 0 || len(r.session.ToolEvents) > 0) {
			sessions = append(sessions, r.session)
		}
	}

	return sessions, nil
}
