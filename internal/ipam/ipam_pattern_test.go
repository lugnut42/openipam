package ipam

import (
	"testing"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPatternBasicFunctions(t *testing.T) {
	// Create a test config with pre-populated patterns
	cfg := &config.Config{
		Patterns: make(map[string]map[string]config.Pattern),
	}
	cfg.Patterns["default"] = make(map[string]config.Pattern)
	
	// Add a test pattern
	cfg.Patterns["default"]["test-pattern"] = config.Pattern{
		CIDRSize:    24,
		Environment: "dev",
		Region:      "us-west1",
		Block:       "10.0.0.0/16",
	}
	
	// Test ListPatterns function with existing patterns
	t.Run("ListPatterns_Success", func(t *testing.T) {
		err := ListPatterns(cfg, "default")
		assert.NoError(t, err)
	})
	
	// Test ListPatterns with non-existent file key
	t.Run("ListPatterns_NonExistentFile", func(t *testing.T) {
		err := ListPatterns(cfg, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no patterns found")
	})
	
	// Test ShowPattern with existing pattern
	t.Run("ShowPattern_Success", func(t *testing.T) {
		err := ShowPattern(cfg, "test-pattern", "default")
		assert.NoError(t, err)
	})
	
	// Test ShowPattern with non-existent pattern
	t.Run("ShowPattern_NonExistentPattern", func(t *testing.T) {
		err := ShowPattern(cfg, "non-existent", "default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}