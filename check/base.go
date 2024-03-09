// Copyright 2023 OWASP ModSecurity Core Rule Set Project
// SPDX-License-Identifier: Apache-2.0

package check

import (
	"bytes"

	"github.com/coreruleset/go-ftw/config"
	"github.com/coreruleset/go-ftw/test"
	"github.com/coreruleset/go-ftw/waflog"
)

// FTWCheck is the base struct for checking test results
type FTWCheck struct {
	log      *waflog.FTWLogLines
	expected *test.Output
	cfg      *config.FTWConfiguration
}

// NewCheck creates a new FTWCheck, allowing to inject the configuration
func NewCheck(c *config.FTWConfiguration) (*FTWCheck, error) {
	ll, err := waflog.NewFTWLogLines(c)
	if err != nil {
		return nil, err
	}
	check := &FTWCheck{
		log:      ll,
		cfg:      c,
		expected: &test.Output{},
	}

	return check, nil
}

// SetExpectTestOutput sets the combined expected output from this test
func (c *FTWCheck) SetExpectTestOutput(t *test.Output) {
	c.expected = t
}

// SetExpectStatus sets to expect the HTTP status from the test to be in the integer range passed
func (c *FTWCheck) SetExpectStatus(status int) {
	c.expected.Status = status
}

// SetExpectResponse sets the response we expect in the text from the server
func (c *FTWCheck) SetExpectResponse(response string) {
	c.expected.ResponseContains = response
}

// SetExpectError sets the boolean if we are expecting an error from the server
func (c *FTWCheck) SetExpectError(expect bool) {
	c.expected.ExpectError = &expect
}

// SetLogContains sets the string to look for in logs
func (c *FTWCheck) SetLogContains(regex string) {
	c.expected.LogContains = regex
	c.expected.Log.MatchRegex = regex
}

// SetNoLogContains sets the string to look that should not present in logs
func (c *FTWCheck) SetNoLogContains(regex string) {
	c.expected.NoLogContains = regex
	c.expected.Log.NoMatchRegex = regex
}

// ForcedIgnore check if this id need to be ignored from results
func (c *FTWCheck) ForcedIgnore(id string) bool {
	for re := range c.cfg.TestOverride.Ignore {
		if re.MatchString(id) {
			return true
		}
	}
	return false
}

// ForcedPass check if this id need to be ignored from results
func (c *FTWCheck) ForcedPass(id string) bool {
	for re := range c.cfg.TestOverride.ForcePass {
		if re.MatchString(id) {
			return true
		}
	}
	return false
}

// ForcedFail check if this id need to be ignored from results
func (c *FTWCheck) ForcedFail(id string) bool {
	for re := range c.cfg.TestOverride.ForceFail {
		if re.MatchString(id) {
			return true
		}
	}
	return false
}

// CloudMode returns true if we are running in cloud mode
func (c *FTWCheck) CloudMode() bool {
	return c.cfg.RunMode == config.CloudRunMode
}

// SetStartMarker sets the log line that marks the start of the logs to analyze
func (c *FTWCheck) SetStartMarker(marker []byte) {
	c.log.StartMarker = bytes.ToLower(marker)
}

// SetEndMarker sets the log line that marks the end of the logs to analyze
func (c *FTWCheck) SetEndMarker(marker []byte) {
	c.log.EndMarker = bytes.ToLower(marker)
}
