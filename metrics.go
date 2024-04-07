package main

import (
	"context"
	"os"
	"time"

	"github.com/go-kit/kit/metrics/graphite"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func NewStatsDClient(
	ctx context.Context,
	prefix string,
	statsdAddress string,
	flushInterval time.Duration,
) *graphite.Graphite {
	logger := log.NewLogfmtLogger(os.Stderr)
	level.NewFilter(logger, level.AllowAll())

	reporter := graphite.New(prefix, logger)
	go reporter.SendLoop(ctx, time.NewTicker(flushInterval).C, "tcp", statsdAddress)

	return reporter
}
