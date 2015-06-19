package honeybadger

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"net/url"
	"os"
	"regexp"
	"time"
)

type hash map[string]interface{}

type Notice struct {
	APIKey       string
	Error        Error
	Token        string
	ErrorMessage string
	ErrorClass   string
	Hostname     string
	Env          string
	Backtrace    []*Frame
	ProjectRoot  string
	Context      Context
	Params       Params
	CGIData      CGIData
	URL          string
}

func (n *Notice) asJSON() *hash {
	return &hash{
		"api_key": n.APIKey,
		"notifier": &hash{
			"name":    "honeybadger",
			"url":     "https://github.com/honeybadger-io/honeybadger-go",
			"version": "0.0.0",
		},
		"error": &hash{
			"token":     n.Token,
			"message":   n.ErrorMessage,
			"class":     n.ErrorClass,
			"backtrace": n.Backtrace,
		},
		"request": &hash{
			"context":  n.Context,
			"params":   n.Params,
			"cgi_data": n.CGIData,
			"url":      n.URL,
		},
		"server": &hash{
			"project_root":     n.ProjectRoot,
			"environment_name": n.Env,
			"hostname":         n.Hostname,
			"time":             time.Now().UTC(),
			"pid":              os.Getpid(),
			"stats":            getStats(),
		},
	}
}

func bytesToKB(bytes uint64) float64 {
	return float64(bytes) / 1024.0
}

func getStats() *hash {
	var m, l *hash

	if stat, err := mem.VirtualMemory(); err == nil {
		m = &hash{
			"total":      bytesToKB(stat.Total),
			"free":       bytesToKB(stat.Free),
			"buffers":    bytesToKB(stat.Buffers),
			"cached":     bytesToKB(stat.Cached),
			"free_total": bytesToKB(stat.Free + stat.Buffers + stat.Cached),
		}
	}

	if stat, err := load.LoadAvg(); err == nil {
		l = &hash{
			"one":     stat.Load1,
			"five":    stat.Load5,
			"fifteen": stat.Load15,
		}
	}

	return &hash{"mem": m, "load": l}
}

func (n *Notice) toJSON() []byte {
	if out, err := json.Marshal(n.asJSON()); err == nil {
		return out
	} else {
		panic(err)
	}
}

func (n *Notice) SetContext(context Context) {
	n.Context.Update(context)
}

func composeStack(stack []*Frame, root string) (frames []*Frame) {
	if root == "" {
		return stack
	}

	re, err := regexp.Compile("^" + regexp.QuoteMeta(root))
	if err != nil {
		return stack
	}

	for _, frame := range stack {
		file := re.ReplaceAllString(frame.File, "[PROJECT_ROOT]")
		frames = append(frames, &Frame{
			File:   file,
			Number: frame.Number,
			Method: frame.Method,
		})
	}
	return
}

func newNotice(config *Configuration, err Error, extra ...interface{}) *Notice {
	notice := Notice{
		APIKey:       config.APIKey,
		Error:        err,
		Token:        uuid.NewRandom().String(),
		ErrorMessage: err.Message,
		ErrorClass:   err.Class,
		Env:          config.Env,
		Hostname:     config.Hostname,
		Backtrace:    composeStack(err.Stack, config.Root),
		ProjectRoot:  config.Root,
		Context:      Context{},
	}

	for _, thing := range extra {
		switch thing := thing.(type) {
		case Context:
			notice.SetContext(thing)
		case Params:
			notice.Params = thing
		case CGIData:
			notice.CGIData = thing
		case url.URL:
			notice.URL = thing.String()
		}
	}

	return &notice
}
