package main

import (
	"net"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	logrusagent "github.com/tengattack/logrus-agent-hook"
	"github.com/victorcoder/dkron/dkron"
	"github.com/victorcoder/dkron/plugin"
)

const (
	// AppID agent app_id
	AppID = "dkron"
)

type AgentOut struct {
	Host       string
	InstanceID string
	loggers    map[string]*log.Logger
	forward    bool
}

func (l *AgentOut) initLogger(dsn string) (*log.Logger, error) {
	logger := log.New()

	// configure log agent (logstash) hook
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		return nil, err
	}
	hook := logrusagent.New(
		conn, logrusagent.DefaultFormatter(
			log.Fields{
				"app_id":      AppID,
				"host":        l.Host,
				"instance_id": l.InstanceID,
			}))
	logger.Hooks.Add(hook)

	return logger, nil
}

func (l *AgentOut) Process(args *dkron.ExecutionProcessorArgs) dkron.Execution {
	logger, dsn := l.parseConfig(args.Config)

	if logger != nil {
		entry := logger.WithTime(args.Execution.StartedAt).WithFields(log.Fields{
			"category": args.Execution.JobName,
			"node":     args.Execution.NodeName,
			"group":    args.Execution.GetGroup(),
		})
		if args.Execution.Success {
			entry.Info(string(args.Execution.Output))
		} else {
			entry.Error(string(args.Execution.Output))
		}
	}

	if !l.forward {
		args.Execution.Output = []byte(dsn)
	}

	return args.Execution
}

func (l *AgentOut) parseConfig(config dkron.PluginConfig) (*log.Logger, string) {
	forward, ok := config["forward"].(bool)
	if ok {
		l.forward = forward
		log.Debugf("Forwarding set to: %t", forward)
	} else {
		l.forward = false
		log.WithField("param", "forward").Warning("Incorrect format or param not found.")
	}

	dsn, ok := config["dsn"].(string)
	if !ok || dsn == "" {
		log.WithField("param", "dsn").Warning("Incorrect format or param not found.")
		return nil, ""
	}

	logger, ok := l.loggers[dsn]
	if !ok {
		var err error
		logger, err = l.initLogger(dsn)
		if err != nil {
			log.WithField("err", err.Error()).Error("Init logger failed.")
			return nil, dsn
		}
		l.loggers[dsn] = logger
	}

	return logger, dsn
}

func main() {
	l := new(AgentOut)
	l.loggers = map[string]*log.Logger{}

	// get default host and instance id from environment variables or hostname
	hostname, _ := os.Hostname()

	// host
	host := os.Getenv("HOST")
	if host == "" {
		host = hostname
	}
	l.Host = host

	// instance_id
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = hostname
	}
	l.InstanceID = instanceID

	plugin.Serve(&plugin.ServeOpts{
		Processor: l,
	})
}
