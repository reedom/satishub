package satis

import (
	"encoding/json"
	"io/ioutil"

	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

// UpdateConfig updates the satis configuration entries.
func UpdateConfig(configPath string, updates []PackageInfo) error {
	var config map[string]interface{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.Errorf("failed to open satis config file: %s", err.Error())
	}

	err = jsoniter.Unmarshal(data, &config)
	if err != nil {
		return errors.Errorf("satis config file contains invalid JSON content: %s", err.Error())
	}

	repos, err := configReadRepos(config)
	if err != nil {
		return err
	}

	requires, err := configReadRequires(config)
	if err != nil {
		return err
	}

	for _, u := range updates {
		found := false
		for _, repo := range repos {
			if repo["url"] == u.URL {
				repo["type"] = u.Type
				found = true
				break
			}
		}

		if !found {
			repos = append(repos, map[string]interface{}{
				"url":  u.URL,
				"type": u.Type,
			})
		}

		if u.Version != "" {
			requires[u.Name] = u.Version
		}
	}

	config["repositories"] = repos
	if 0 < len(requires) {
		config["require"] = requires
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Errorf("failed to encode satis config file: %s", err)
	}

	// FIXME use temporary file to prevent possibly making a broken file when disk got full
	err = ioutil.WriteFile(configPath, data, 0644)
	if err != nil {
		return errors.Errorf("failed to write satis config file: %s", err)
	}

	return nil
}

func configReadRepos(config map[string]interface{}) ([]map[string]interface{}, error) {
	tmp, ok := config["repositories"]
	if !ok {
		return make([]map[string]interface{}, 0), nil
	}

	tmp2, ok := tmp.([]interface{})
	if !ok {
		return nil, errors.Errorf(`config entry "repository" is not an array`)
	}

	repos := make([]map[string]interface{}, len(tmp2))
	for i, repo := range tmp2 {
		repos[i], ok = repo.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf(`config entry "repository[%d]" is not a hash`, i)
		}
	}

	return repos, nil
}

func configReadRequires(config map[string]interface{}) (map[string]interface{}, error) {
	tmp, ok := config["require"]
	if !ok {
		return make(map[string]interface{}, 0), nil
	}

	requires, ok := tmp.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf(`config entry "require" is not a hash`)
	}
	return requires, nil
}
