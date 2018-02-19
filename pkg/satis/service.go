package satis

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gin-gonic/gin/json"
	"github.com/pkg/errors"
)

// Service represents satishub servies.
type Service interface {
	Close()
	Run(ctx context.Context) <-chan ServiceResult
	Rebuild() chan ServiceResult
	UpdatePackage(pkg PackageInfo) chan ServiceResult

	ConfigPath() string
	RepoPath() string
}

type requestPartial struct {
	PackageInfo
	Result chan ServiceResult
}

// service represents statis service.
type service struct {
	satisPath   string
	configPath  string
	repoPath    string
	debug       bool
	timeout     time.Duration
	snsTopicARN string
	errLog      *log.Logger
	stdLog      *log.Logger

	cmdRebuild chan chan ServiceResult
	cmdPartial chan requestPartial
	closeOnce  sync.Once
}

// ServiceParam contains parameters to NewService() call.
type ServiceParam struct {
	SatisPath   string
	ConfigPath  string
	RepoPath    string
	Debug       bool
	Timeout     time.Duration
	ErrLog      *log.Logger
	StdLog      *log.Logger
	SNSTopicARN string
}

// NewService creates service instance with the specified parameters.
func NewService(param ServiceParam) Service {
	s := &service{
		satisPath:   param.SatisPath,
		configPath:  param.ConfigPath,
		repoPath:    param.RepoPath,
		debug:       param.Debug,
		timeout:     param.Timeout,
		snsTopicARN: param.SNSTopicARN,
		errLog:      param.ErrLog,
		stdLog:      param.StdLog,
		cmdRebuild:  make(chan chan ServiceResult, 16),
		cmdPartial:  make(chan requestPartial, 16),
	}

	if s.errLog == nil {
		s.errLog = log.New(os.Stdout, "satis ", log.Ldate|log.Ltime)
	}
	if s.stdLog == nil {
		s.stdLog = log.New(os.Stdout, "satis ", log.Ldate|log.Ltime)

	}
	return s
}

// Close closes the service.
func (s *service) Close() {
	s.closeOnce.Do(func() {
		close(s.cmdRebuild)
		close(s.cmdPartial)
	})
}

// ConfigPath returns satis config file path.
func (s *service) ConfigPath() string {
	return s.configPath
}

// RepoPath returns repository path (to where satis outputs).
func (s *service) RepoPath() string {
	return s.repoPath
}

// Run starts the service.
func (s *service) Run(ctx context.Context) <-chan ServiceResult {
	result := make(chan ServiceResult)

	go func() {
		s.notifyService("satishub service start")
		defer func() {
			if s.debug {
				s.stdLog.Print("service close")
			}
			s.discardCommands()
			close(result)
			s.notifyService("satishub service exit!")
		}()

		for {
			if s.debug {
				s.stdLog.Print("wait for command...")
			}
			select {
			case <-ctx.Done():
				return
			case ch, ok := <-s.cmdRebuild:
				if !ok {
					return
				}
				if s.debug {
					s.stdLog.Println("cmd rebuild")
				}
				s.discardCommands()
				// TODO make it possible to cancel previous Execute command
				ctxCmd, cancel := context.WithTimeout(ctx, s.timeout)
				err := s.rebuild(ctxCmd)
				if err == context.DeadlineExceeded {
					err = errors.Wrap(err, "satis command execution timeout")
				}
				cancel()
				r := ServiceResult{Error: err}
				result <- r
				ch <- r
				close(ch)
			case req, ok := <-s.cmdPartial:
				if !ok {
					return
				}
				if s.debug {
					s.stdLog.Println("cmd partial build")
				}
				err := s.notifyPartialBuild(req.PackageInfo, "start", nil)
				if err != nil {
					s.errLog.Println("notify error", err.Error())
				}
				err = s.updatePackage(ctx, req.PackageInfo)
				if err != nil {
					s.notifyPartialBuild(req.PackageInfo, "error", err)
					if err == context.DeadlineExceeded {
						err = errors.Wrap(err, "satis command execution timeout")
					}
				} else {
					err = s.notifyPartialBuild(req.PackageInfo, "completed", nil)
					if err != nil {
						s.errLog.Println("notify error", err.Error())
					}
				}
				r := ServiceResult{Error: err}
				result <- r
				req.Result <- r
				close(req.Result)
			}
		}
	}()
	return result
}

type snsTopicPartial struct {
	Event   string      `json:"type"`
	Package PackageInfo `json:"package"`
	Msg     string      `json:"msg"`
	Error   string      `json:"error,omitempty"`
	Time    int64       `json:"time"`
}

func (s *service) notifyPartialBuild(info PackageInfo, msg string, serviceErr error) error {
	if s.snsTopicARN == "" {
		return nil
	}

	payload := snsTopicPartial{
		Event:   "partialBuild",
		Package: info,
		Msg:     msg,
		Time:    time.Now().Unix(),
	}
	if serviceErr != nil {
		payload.Error = serviceErr.Error()
	}

	data, _ := json.Marshal(payload)
	return Notify(s.snsTopicARN, string(data))
}

type snsTopicService struct {
	Event string `json:"type"`
	Msg   string `json:"msg"`
	Time  int64  `json:"time"`
}

func (s *service) notifyService(msg string) error {
	if s.snsTopicARN == "" {
		return nil
	}

	payload := snsTopicService{
		Event: "service",
		Msg:   msg,
		Time:  time.Now().Unix(),
	}

	data, _ := json.Marshal(payload)
	return Notify(s.snsTopicARN, string(data))
}

// Rebuild requests satis full rebuild.
func (s *service) Rebuild() chan ServiceResult {
	ch := make(chan ServiceResult)
	s.cmdRebuild <- ch
	return ch
}

// UpdatePackage requests updating the satis config file and partial building.
func (s *service) UpdatePackage(pkg PackageInfo) chan ServiceResult {
	ch := make(chan ServiceResult)
	s.cmdPartial <- requestPartial{pkg, ch}
	return ch
}

func (s *service) updatePackage(ctx context.Context, pkg PackageInfo) error {
	err := UpdateConfig(s.configPath, []PackageInfo{pkg})
	if err != nil {
		return err
	}

	ctxCmd, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if pkg.Name != "" {
		return s.partialBuild(ctxCmd, pkg.Name)
	}
	return s.rebuild(ctxCmd)
}

func (s *service) discardCommands() {
	for {
		select {
		case ch, ok := <-s.cmdRebuild:
			if ok {
				close(ch)
			}
		case cmd, ok := <-s.cmdPartial:
			if ok {
				close(cmd.Result)
			}
		default:
			return
		}
	}
}

func (s *service) rebuild(ctx context.Context) error {
	command := exec.CommandContext(ctx, s.satisPath, "build", s.configPath, s.repoPath)
	command.Stdout = logWriter{s.stdLog}
	command.Stderr = logWriter{s.errLog}
	return command.Run()
}

func (s *service) partialBuild(ctx context.Context, targetPackage string) error {
	command := exec.CommandContext(ctx, s.satisPath, "build", s.configPath, s.repoPath, targetPackage)
	command.Stdout = logWriter{s.stdLog}
	command.Stderr = logWriter{s.errLog}
	return command.Run()
}

type logWriter struct {
	logger *log.Logger
}

func (w logWriter) Write(data []byte) (int, error) {
	w.logger.Print(string(data))
	return len(data), nil
}
