/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package openapi_cases

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/flant/addon-operator/pkg/utils"
	"github.com/flant/addon-operator/pkg/values/validation"
	"sigs.k8s.io/yaml"
)

const FocusFieldName = "x-test-focus"

type TestCase struct {
	ConfigValues []map[string]interface{}
	Values       []map[string]interface{}
	HelmValues   []map[string]interface{}
}

func (tc TestCase) HasFocused() bool {
	for _, values := range tc.ConfigValues {
		if _, hasFocus := values[FocusFieldName]; hasFocus {
			return true
		}
	}
	for _, values := range tc.Values {
		if _, hasFocus := values[FocusFieldName]; hasFocus {
			return true
		}
	}
	for _, values := range tc.HelmValues {
		if _, hasFocus := values[FocusFieldName]; hasFocus {
			return true
		}
	}
	return false
}

type TestCases struct {
	Positive TestCase
	Negative TestCase

	dir        string
	moduleName string
	hasFocused bool
}

func (t *TestCases) HaveConfigValuesCases() bool {
	return len(t.Positive.ConfigValues) > 0 || len(t.Negative.ConfigValues) > 0
}

func (t *TestCases) HaveValuesCases() bool {
	return len(t.Positive.Values) > 0 || len(t.Negative.Values) > 0
}

func (t *TestCases) HaveHelmValuesCases() bool {
	return len(t.Positive.HelmValues) > 0 || len(t.Negative.HelmValues) > 0
}

func GetAllOpenAPIDirs() ([]string, error) {
	var (
		dirs        []string
		openAPIDirs []string
	)

	for _, possibleDir := range []string{
		"/deckhouse/modules/*/openapi",
		"/deckhouse/ee/modules/*/openapi",
		"/deckhouse/ee/fe/modules/*/openapi",
	} {
		globDirs, err := filepath.Glob(possibleDir)
		if err != nil {
			return nil, err
		}

		openAPIDirs = append(openAPIDirs, globDirs...)
	}

	openAPIDirs = append(openAPIDirs, "/deckhouse/global-hooks/openapi")
	for _, openAPIDir := range openAPIDirs {
		info, err := os.Stat(openAPIDir)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			continue
		}
		dirs = append(dirs, openAPIDir)
	}
	return dirs, nil
}

func TestCasesFromFile(filename string) (*TestCases, error) {
	var testCases TestCases
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &testCases)
	if err != nil {
		return nil, err
	}
	testCases.hasFocused = testCases.Positive.HasFocused() || testCases.Negative.HasFocused()
	return &testCases, nil
}

func ValidatePositiveCase(validator *validation.ValuesValidator, moduleName string, schema validation.SchemaType, testValues map[string]interface{}, runFocused bool) error {
	if _, hasFocus := testValues[FocusFieldName]; !hasFocus && runFocused {
		return nil
	}
	delete(testValues, FocusFieldName)
	return validator.ValidateValues(validation.ModuleSchema, schema, moduleName, utils.Values{moduleName: testValues})
}

func ValidateNegativeCase(validator *validation.ValuesValidator, moduleName string, schema validation.SchemaType, testValues map[string]interface{}, runFocused bool) error {
	_, hasFocus := testValues[FocusFieldName]
	if !hasFocus && runFocused {
		return nil
	}
	delete(testValues, FocusFieldName)
	err := validator.ValidateValues(validation.ModuleSchema, schema, moduleName, utils.Values{moduleName: testValues})
	if err == nil {
		return fmt.Errorf("negative case error for %s values: test case should not pass validation: %+v", schema, ValuesToString(testValues))
	}
	// Focusing is a debugging tool, so print hidden error.
	if hasFocus {
		fmt.Printf("Debug: expected error for negative case: %v\n", err)
	}
	return nil
}

func ValuesToString(v map[string]interface{}) string {
	b, _ := yaml.Marshal(v)
	return string(b)
}
