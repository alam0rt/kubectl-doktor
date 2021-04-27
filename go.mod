module github.com/alam0rt/kubectl-doktor

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.21.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/ucloud/ucloud-sdk-go v0.21.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.0
)
