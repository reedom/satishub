package satis_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/reedom/satishub/pkg/satis"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	config, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	defer func() {
		assert.NoError(t, os.Remove(config.Name()))
	}()
	config.WriteString("{}")
	assert.NoError(t, config.Close())

	stdBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	param := satis.ServiceParam{
		SatisPath:  "echo",
		ConfigPath: config.Name(),
		RepoPath:   "outRepoDir",
		Debug:      false,
		Timeout:    5 * time.Second,
		ErrLog:     log.New(errBuf, "satis err ", 0),
		StdLog:     log.New(stdBuf, "satis ", 0),
	}
	s := satis.NewService(param)
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := s.Run(ctx)

	// UpdatePackage
	pkg := satis.PackageInfo{
		Name: "test/pkg",
		URL:  "http://example.com/pkg",
		Type: "vcs",
	}

	ch2 := s.UpdatePackage(pkg)
	select {
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	case result := <-ch:
		assert.NoError(t, result.Error)
	}

	select {
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	case result := <-ch2:
		assert.NoError(t, result.Error)
	}

	assert.Empty(t, string(errBuf.Bytes()))
	expected := fmt.Sprintf("satis build %v %v %v\n", param.ConfigPath, param.RepoPath, pkg.Name)
	assert.Equal(t, expected, string(stdBuf.Bytes()))

	// Rebuild

	ch2 = s.Rebuild()
	select {
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	case result := <-ch:
		assert.NoError(t, result.Error)
	}

	select {
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	case result := <-ch2:
		assert.NoError(t, result.Error)
	}

	expected += fmt.Sprintf("satis build %v %v\n", param.ConfigPath, param.RepoPath)
	assert.Equal(t, expected, string(stdBuf.Bytes()))

	cancel()

	// wait for cancel() affects
	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "close timeout")
	}

	// wait for Service.Run() ends
	select {
	case <-ch:
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "end timeout")
	}
}

func createServer(t *testing.T, handler func(ctx context.Context, s satis.Service, ch <-chan satis.ServiceResult, wout, werr io.Writer)) {
	config, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
		return
	}
	defer func() {
		assert.NoError(t, os.Remove(config.Name()))
	}()
	config.WriteString("{}")
	assert.NoError(t, config.Close())

	stdBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	param := satis.ServiceParam{
		SatisPath:  "echo",
		ConfigPath: config.Name(),
		RepoPath:   "outRepoDir",
		Debug:      false,
		Timeout:    0,
		ErrLog:     log.New(errBuf, "satis err ", 0),
		StdLog:     log.New(stdBuf, "satis ", 0),
	}
	s := satis.NewService(param)
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := s.Run(ctx)
	handler(ctx, s, ch, stdBuf, errBuf)
}

func TestUpdatePackageTimeout(t *testing.T) {
	createServer(t, func(ctx context.Context, s satis.Service, ch <-chan satis.ServiceResult, wout, werr io.Writer) {
		// UpdatePackage
		pkg := satis.PackageInfo{
			Name: "test/pkg",
			URL:  "http://example.com/pkg",
			Type: "vcs",
		}

		s.UpdatePackage(pkg)
		select {
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		case result := <-ch:
			assert.Contains(t, result.Error.Error(), "satis command execution timeout")
		}
	})
}

func TestRebuildTimeout(t *testing.T) {
	createServer(t, func(ctx context.Context, s satis.Service, ch <-chan satis.ServiceResult, wout, werr io.Writer) {
		s.Rebuild()
		select {
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		case result := <-ch:
			assert.Contains(t, result.Error.Error(), "satis command execution timeout")
		}
	})
}
