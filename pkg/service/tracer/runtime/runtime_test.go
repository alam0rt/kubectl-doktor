package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainerRuntimeBridge_Docker(t *testing.T) {
	bridge := NewContainerRuntimeBridge("docker")
	assert.IsType(t, &DockerBridge{}, bridge)
}

func TestNewContainerRuntimeBridge_Invalid(t *testing.T) {
	assert.Panics(t, func() { NewContainerRuntimeBridge("i-do-not-exist") })
}
