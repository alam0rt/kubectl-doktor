package tracer

import (
	"bytes"
	"io"

	"github.com/alam0rt/kubectl-doktor/kube"
	"github.com/alam0rt/kubectl-doktor/pkg/config"
	"github.com/alam0rt/kubectl-doktor/pkg/service/tracer/runtime"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

type PrivilegedPodTracerService struct {
	settings                *config.DoktorSettings
	privilegedPod           *v1.Pod
	privilegedContainerName string
	targetProcessId         *string
	kubernetesApiService    kube.KubernetesApiService
	runtimeBridge           runtime.ContainerRuntimeBridge
}

func NewPrivilegedPodRemoteTracingService(options *config.DoktorSettings, service kube.KubernetesApiService, bridge runtime.ContainerRuntimeBridge) TracerService {
	return &PrivilegedPodTracerService{settings: options, privilegedContainerName: "doktor-privileged", kubernetesApiService: service, runtimeBridge: bridge}
}

func (p *PrivilegedPodTracerService) Setup() error {
	var err error

	log.Info().
		Msgf("creating privileged pod on node: '%s'", p.settings.DetectedPodNodeName)

	if p.settings.UseDefaultImage {
		p.settings.Image = p.runtimeBridge.GetDefaultImage()
	}

	if p.settings.UseDefaultSocketPath {
		p.settings.SocketPath = p.runtimeBridge.GetDefaultSocketPath()
	}

	p.privilegedPod, err = p.kubernetesApiService.CreatePrivilegedPod(
		p.settings.DetectedPodNodeName,
		p.privilegedContainerName,
		p.settings.Image,
		p.settings.SocketPath,
		p.settings.UserSpecifiedPodCreateTimeout,
	)
	if err != nil {
		log.Error().
			Msgf("failed to create privileged pod on node: '%s'", p.settings.DetectedPodNodeName)
		return err
	}

	log.Info().
		Msgf("pod: '%s' created successfully on node: '%s'", p.privilegedPod.Name, p.settings.DetectedPodNodeName)

	if p.runtimeBridge.NeedsPid() {
		var buff bytes.Buffer
		command := p.runtimeBridge.BuildInspectCommand(p.settings.DetectedContainerId)
		exitCode, err := p.kubernetesApiService.ExecuteCommand(p.privilegedPod.Name, p.privilegedContainerName, command, &buff)
		if err != nil {
			log.Error().
				Msgf("failed to start tracing using privileged pod, exit code: '%d'", exitCode)
		}
		p.targetProcessId, err = p.runtimeBridge.ExtractPid(buff.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PrivilegedPodTracerService) Cleanup() error {
	log.Info().
		Msgf("removing privileged container: '%s'", p.privilegedContainerName)

	command := p.runtimeBridge.BuildCleanupCommand()

	exitCode, err := p.kubernetesApiService.ExecuteCommand(p.privilegedPod.Name, p.privilegedContainerName, command, &kube.NopWriter{})
	if err != nil {
		log.Error().
			Msgf("failed to remove privileged container: '%s', exit code: '%d', "+
				"please manually remove it", p.privilegedContainerName, exitCode)
	} else {
		log.Info().
			Msgf("privileged container: '%s' removed successfully", p.privilegedContainerName)
	}

	log.Info().
		Msgf("removing pod: '%s'", p.privilegedPod.Name)

	err = p.kubernetesApiService.DeletePod(p.privilegedPod.Name)
	if err != nil {
		log.Error().
			Msgf("failed to remove pod: '%s", p.privilegedPod.Name)
		return err
	}

	log.Info().
		Msgf("pod: '%s' removed successfully", p.privilegedPod.Name)

	return nil
}

func (p *PrivilegedPodTracerService) Start(stdOut io.Writer) error {
	log.Info().
		Msgf("starting remote tracing using privileged pod")

	command := p.runtimeBridge.BuildTestCommand()

	exitCode, err := p.kubernetesApiService.ExecuteCommand(p.privilegedPod.Name, p.privilegedContainerName, command, stdOut)
	if err != nil {
		log.Error().
			Msgf("failed to start tracing using privileged pod, exit code: '%d'", exitCode)
		return err
	}

	log.Info().
		Msg("remote tracing using privileged pod completed")

	return nil
}
