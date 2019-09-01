package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg, err := FromYAMLFile("./example_filesystem.yml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}
