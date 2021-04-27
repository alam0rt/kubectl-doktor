package runtime

import "fmt"

var SupportedContainerRuntimes = []string{
	"docker",
}

type ContainerRuntimeBridge interface {
	NeedsPid() bool
	BuildInspectCommand(containerId string) []string
	ExtractPid(inspection string) (*string, error)
	BuildTcpdumpCommand(containerId *string, netInterface string, filter string, pid *string, socketPath string) []string
	BuildTestCommand() []string
	BuildCleanupCommand() []string
	GetDefaultImage() string
	GetDefaultSocketPath() string
}

func NewContainerRuntimeBridge(runtimeName string) ContainerRuntimeBridge {
	switch runtimeName {
	case "docker":
		return NewDockerBridge()
	default:
		panic(fmt.Sprintf("Unable to build bridge to %s", runtimeName))
	}
}
