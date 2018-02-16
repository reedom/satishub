package satis_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/reedom/satishub/pkg/satis"
	"github.com/stretchr/testify/assert"
)

func TestUpdateEmptyConfig(t *testing.T) {
	// create an empty file as a config
	tmp, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	tmp.WriteString(`{}`)
	tmp.Close()
	defer os.Remove(tmp.Name())

	updates := []satis.PackageInfo{satis.PackageInfo{
		Name: "test/pkg",
		URL:  "http://example.com/pkg",
		Type: "vcs",
	}}
	assert.NoError(t, satis.UpdateConfig(tmp.Name(), updates))

	expected := `{
  "repositories": [
    {
      "type": "vcs",
      "url": "http://example.com/pkg"
    }
  ]
}`
	config, err := ioutil.ReadFile(tmp.Name())
	assert.NoError(t, err)
	assert.Equal(t, expected, string(config))
}

func TestUpdateConfig(t *testing.T) {
	content := `{
  "repositories": [
    {
      "type": "vcs",
      "url": "http://example.com/pkg"
    }
  ]
}`

	tmp, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	tmp.WriteString(content)
	tmp.Close()
	defer os.Remove(tmp.Name())

	updates := []satis.PackageInfo{satis.PackageInfo{
		Name:    "test/another-pkg",
		URL:     "http://example.com/another-pkg",
		Type:    "something",
		Version: "1.0.2",
	}}
	assert.NoError(t, satis.UpdateConfig(tmp.Name(), updates))

	expected := `{
  "repositories": [
    {
      "type": "vcs",
      "url": "http://example.com/pkg"
    },
    {
      "type": "something",
      "url": "http://example.com/another-pkg"
    }
  ],
  "require": {
    "test/another-pkg": "1.0.2"
  }
}`
	config, err := ioutil.ReadFile(tmp.Name())
	assert.NoError(t, err)
	assert.Equal(t, expected, string(config))
}

func TestConfigNotFound(t *testing.T) {
	err := satis.UpdateConfig("/nowhere", []satis.PackageInfo{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open satis config file")
}

func TestInvalidConfigJSON(t *testing.T) {
	tmp, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	tmp.WriteString(`}`)
	tmp.Close()
	defer os.Remove(tmp.Name())

	err = satis.UpdateConfig(tmp.Name(), []satis.PackageInfo{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "satis config file contains invalid JSON content: ")
}

func TestInvalidConfigJSON2(t *testing.T) {
	tmp, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	tmp.WriteString(`{"repositories":0}`)
	tmp.Close()
	defer os.Remove(tmp.Name())

	err = satis.UpdateConfig(tmp.Name(), []satis.PackageInfo{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `config entry "repository" is not an array`)
}

func TestInvalidConfigJSON3(t *testing.T) {
	tmp, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	tmp.WriteString(`{"repositories":[{},1]}`)
	tmp.Close()
	defer os.Remove(tmp.Name())

	err = satis.UpdateConfig(tmp.Name(), []satis.PackageInfo{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `config entry "repository[1]" is not a hash`)
}

func TestInvalidConfigJSON4(t *testing.T) {
	tmp, err := ioutil.TempFile("", "satis-test")
	assert.NoError(t, err)
	tmp.WriteString(`{"require":0}`)
	tmp.Close()
	defer os.Remove(tmp.Name())

	err = satis.UpdateConfig(tmp.Name(), []satis.PackageInfo{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `config entry "require" is not a hash`)
}
