package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg, err := FromYAMLFile("./example.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}
