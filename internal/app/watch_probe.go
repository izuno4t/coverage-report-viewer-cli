package app

import (
	"fmt"
	"os"
	"time"
)

type fileStamp struct {
	modTime time.Time
	size    int64
}

type reportUpdateProbe struct {
	paths []string
	last  map[string]fileStamp
}

func newReportUpdateProbe(paths []string) (func() (bool, error), error) {
	probe := &reportUpdateProbe{
		paths: append([]string(nil), paths...),
		last:  make(map[string]fileStamp, len(paths)),
	}
	if err := probe.captureInitial(); err != nil {
		return nil, err
	}
	return probe.HasChanged, nil
}

func (p *reportUpdateProbe) captureInitial() error {
	for _, path := range p.paths {
		stamp, err := statFileStamp(path)
		if err != nil {
			return err
		}
		p.last[path] = stamp
	}
	return nil
}

func (p *reportUpdateProbe) HasChanged() (bool, error) {
	for _, path := range p.paths {
		current, err := statFileStamp(path)
		if err != nil {
			return false, err
		}
		prev, ok := p.last[path]
		if !ok || prev != current {
			p.last[path] = current
			return true, nil
		}
	}
	return false, nil
}

func statFileStamp(path string) (fileStamp, error) {
	info, err := os.Stat(path)
	if err != nil {
		return fileStamp{}, fmt.Errorf("%s: %w", path, err)
	}
	return fileStamp{
		modTime: info.ModTime(),
		size:    info.Size(),
	}, nil
}
