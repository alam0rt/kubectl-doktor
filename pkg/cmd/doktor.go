package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alam0rt/kubectl-doktor/kube"
	"github.com/alam0rt/kubectl-doktor/pkg/config"
	"github.com/alam0rt/kubectl-doktor/pkg/service/tracer"
	"github.com/alam0rt/kubectl-doktor/pkg/service/tracer/runtime"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

var (
	doktorExample = `
	%[1]s doktor example-pod -n default
	`
)

type Doktor struct {
	configFlags      *genericclioptions.ConfigFlags
	resultingContext *api.Context
	clientset        *kubernetes.Clientset
	restConfig       *rest.Config
	rawConfig        api.Config
	settings         *config.DoktorSettings
	tracerService    tracer.TracerService
}

func NewDoktor(settings *config.DoktorSettings) *Doktor {
	return &Doktor{settings: settings, configFlags: genericclioptions.NewConfigFlags(true)}
}

// NamespaceOptions provides information required to update
// the current context on a user's KUBECONFIG

func NewCmdDoktor(streams genericclioptions.IOStreams) *cobra.Command {
	doktorSettings := config.NewDoktorSettings(streams)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	doktor := NewDoktor(doktorSettings)

	cmd := &cobra.Command{
		Use:          "kubectl doktor",
		Short:        "What'chu wanna know?!",
		Example:      fmt.Sprintf(doktorExample, "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := doktor.Complete(c, args); err != nil {
				return err
			}
			if err := doktor.Validate(); err != nil {
				return err
			}
			if err := doktor.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&doktorSettings.UserSpecifiedNamespace, "namespace", "n", "", "namespace (optional)")
	_ = viper.BindEnv("namespace", "KUBECTL_PLUGINS_CURRENT_NAMESPACE")
	_ = viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))

	cmd.Flags().StringVarP(&doktorSettings.UserSpecifiedContainer, "container", "c", "", "container (optional)")
	_ = viper.BindEnv("container", "KUBECTL_PLUGINS_LOCAL_FLAG_CONTAINER")
	_ = viper.BindPFlag("container", cmd.Flags().Lookup("container"))

	cmd.Flags().StringVarP(&doktorSettings.UserSpecifiedFilter, "filter", "f", "", "bpftrace filter (optional)")
	_ = viper.BindEnv("filter", "KUBECTL_PLUGINS_LOCAL_FLAG_FILTER")
	_ = viper.BindPFlag("filter", cmd.Flags().Lookup("filter"))

	cmd.Flags().BoolVarP(&doktorSettings.UserSpecifiedVerboseMode, "verbose", "v", false,
		"if specified, ksniff output will include debug information (optional)")
	_ = viper.BindEnv("verbose", "KUBECTL_PLUGINS_LOCAL_FLAG_VERBOSE")
	_ = viper.BindPFlag("verbose", cmd.Flags().Lookup("verbose"))

	cmd.Flags().BoolVarP(&doktorSettings.UserSpecifiedPrivilegedMode, "privileged", "p", false,
		"if specified, doktor will deploy another pod that have privileges to attach to host namespace")
	_ = viper.BindEnv("privileged", "KUBECTL_PLUGINS_LOCAL_FLAG_PRIVILEGED")
	_ = viper.BindPFlag("privileged", cmd.Flags().Lookup("privileged"))

	cmd.Flags().DurationVarP(&doktorSettings.UserSpecifiedPodCreateTimeout, "pod-creation-timeout", "",
		1*time.Minute, "the length of time to wait for privileged pod to be created (e.g. 20s, 2m, 1h). "+
			"A value of zero means the creation never times out.")

	cmd.Flags().StringVarP(&doktorSettings.Image, "image", "", "",
		"the privileged container image (optional)")
	_ = viper.BindEnv("image", "KUBECTL_PLUGINS_LOCAL_FLAG_IMAGE")
	_ = viper.BindPFlag("image", cmd.Flags().Lookup("image"))

	cmd.Flags().StringVarP(&doktorSettings.UserSpecifiedKubeContext, "context", "x", "",
		"kubectl context to work on (optional)")
	_ = viper.BindEnv("context", "KUBECTL_PLUGINS_CURRENT_CONTEXT")
	_ = viper.BindPFlag("context", cmd.Flags().Lookup("context"))

	cmd.Flags().StringVarP(&doktorSettings.SocketPath, "socket", "", "",
		"the container runtime socket path (optional)")
	_ = viper.BindEnv("socket", "KUBECTL_PLUGINS_SOCKET_PATH")
	_ = viper.BindPFlag("socket", cmd.Flags().Lookup("socket"))

	return cmd
}

func (o *Doktor) Run() error {
	log.Info().
		Str("pod", o.settings.UserSpecifiedPodName).
		Str("namespace", o.resultingContext.Namespace).
		Str("container", o.settings.UserSpecifiedContainer).
		Str("filter", o.settings.UserSpecifiedFilter).
		Msg("tracing has begun")

	err := o.tracerService.Setup()
	if err != nil {
		return err
	}

	defer func() {
		log.Info().
			Msg("starting sniffer cleanup")

		err := o.tracerService.Cleanup()
		if err != nil {
			log.Error().
				Msg("failed to teardown sniffer, a manual teardown is required.")

			return
		}

		log.Info().
			Msg("sniffer cleanup completed successfully")
	}()

	return nil
}

func (o *Doktor) Complete(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		_ = cmd.Usage()
		return errors.New("provide more arguments")

	}

	o.settings.UserSpecifiedPodName = args[0]
	if o.settings.UserSpecifiedPodName == "" {
		return errors.New("pod name is empty")
	}

	o.settings.UserSpecifiedNamespace = viper.GetString("namespace")
	o.settings.UserSpecifiedContainer = viper.GetString("container")
	o.settings.UserSpecifiedInterface = viper.GetString("interface")
	o.settings.UserSpecifiedFilter = viper.GetString("filter")
	o.settings.UserSpecifiedVerboseMode = viper.GetBool("verbose")
	o.settings.UserSpecifiedPrivilegedMode = viper.GetBool("privileged")
	o.settings.UserSpecifiedKubeContext = viper.GetString("context")
	o.settings.UseDefaultImage = !cmd.Flag("image").Changed
	o.settings.UseDefaultSocketPath = !cmd.Flag("socket").Changed

	var err error

	if o.settings.UserSpecifiedVerboseMode {
		log.Info().
			Msg("running in verbose mode")
	}

	o.rawConfig, err = o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	var currentContext *api.Context
	var exists bool

	if o.settings.UserSpecifiedKubeContext != "" {
		currentContext, exists = o.rawConfig.Contexts[o.settings.UserSpecifiedKubeContext]
	} else {
		currentContext, exists = o.rawConfig.Contexts[o.rawConfig.CurrentContext]
	}

	if !exists {
		return errors.New("context doesn't exist")
	}

	o.restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: o.configFlags.ToRawKubeConfigLoader().ConfigAccess().GetDefaultFilename()},
		&clientcmd.ConfigOverrides{
			CurrentContext: o.settings.UserSpecifiedKubeContext,
		}).ClientConfig()

	if err != nil {
		return err
	}

	o.restConfig.Timeout = 30 * time.Second

	o.clientset, err = kubernetes.NewForConfig(o.restConfig)
	if err != nil {
		return err
	}

	o.resultingContext = currentContext.DeepCopy()
	if o.settings.UserSpecifiedNamespace != "" {
		o.resultingContext.Namespace = o.settings.UserSpecifiedNamespace
	}

	return nil
}

func (o *Doktor) Validate() error {
	if len(o.rawConfig.CurrentContext) == 0 {
		return errors.New("context doesn't exist")
	}

	if o.resultingContext.Namespace == "" {
		return errors.New("namespace value is empty should be custom or default")
	}

	var err error

	pod, err := o.clientset.CoreV1().Pods(o.resultingContext.Namespace).Get(context.TODO(), o.settings.UserSpecifiedPodName, v1.GetOptions{})
	if err != nil {
		return err
	}

	o.settings.DetectedPodNodeName = pod.Spec.NodeName

	log.Debug().
		Str("pod", o.settings.UserSpecifiedPodName).
		Str("phase", string(pod.Status.Phase))

	if len(pod.Spec.Containers) < 1 {
		return errors.New("no containers in specified pod")
	}

	if o.settings.UserSpecifiedContainer == "" {
		log.Info().
			Msg("no container specified, taking first container we found in pod.")

		o.settings.UserSpecifiedContainer = pod.Spec.Containers[0].Name

		log.Info().
			Str("selected container", o.settings.UserSpecifiedContainer)
	}

	if err := o.findContainerId(pod); err != nil {
		return err
	}

	kubernetesApiService := kube.NewKubernetesApiService(o.clientset, o.restConfig, o.resultingContext.Namespace)

	if o.settings.UserSpecifiedPrivilegedMode {
		log.Info().
			Str("sniffing method", "privileged pod")
		bridge := runtime.NewContainerRuntimeBridge(o.settings.DetectedContainerRuntime)
		o.tracerService = tracer.NewPrivilegedPodRemoteTracingService(o.settings, kubernetesApiService, bridge)
		log.Info().
			Str("detected bridge", bridge.GetDefaultImage())
	} else {
		log.Fatal().
			Msg("tracing method not yet implemented")
	}

	return nil
}

func (o *Doktor) findContainerId(pod *corev1.Pod) error {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if o.settings.UserSpecifiedContainer == containerStatus.Name {
			result := strings.Split(containerStatus.ContainerID, "://")
			if len(result) != 2 {
				break
			}
			o.settings.DetectedContainerRuntime = result[0]
			o.settings.DetectedContainerId = result[1]
			return nil
		}
	}

	return errors.Errorf("couldn't find container: '%s' in pod: '%s'", o.settings.UserSpecifiedContainer, o.settings.UserSpecifiedPodName)
}
