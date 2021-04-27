package config

import (
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type DoktorSettings struct {
	UserSpecifiedPodName          string
	UserSpecifiedInterface        string
	UserSpecifiedFilter           string
	UserSpecifiedPodCreateTimeout time.Duration
	UserSpecifiedContainer        string
	UserSpecifiedNamespace        string
	UserSpecifiedVerboseMode      bool
	UserSpecifiedPrivilegedMode   bool
	UserSpecifiedImage            string
	DetectedPodNodeName           string
	DetectedContainerId           string
	DetectedContainerRuntime      string
	Image                         string
	UseDefaultImage               bool
	UserSpecifiedKubeContext      string
	SocketPath                    string
	UseDefaultSocketPath          bool
}

func NewDoktorSettings(streams genericclioptions.IOStreams) *DoktorSettings {
	return &DoktorSettings{}
}
