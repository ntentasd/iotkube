package helm

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
)

type Release struct {
	name       string
	namespace  string
	version    string
	repository string
	values     map[string]any
}

func NewRelease(name, namespace, version, repository string) *Release {
	return &Release{
		name,
		namespace,
		version,
		repository,
		map[string]any{},
	}
}

func (r *Release) ApplyValues(values map[string]any) {
	r.values = values
}

func (r *Release) Install() error {
	settings := cli.New()

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), r.namespace, os.Getenv("HELM_DRIVER"), func(format string, v ...any) {}); err != nil {
		return err
	}

	installClient := action.NewInstall(actionConfig)
	installClient.DryRunOption = "none"
	installClient.ReleaseName = r.name
	installClient.Namespace = r.namespace
	installClient.Version = r.version

	registryClient, err := newRegistryClientTLS(settings, nil, installClient.CertFile, installClient.KeyFile, installClient.CaFile, installClient.InsecureSkipTLSverify, installClient.PlainHTTP)
	if err != nil {
		return err
	}

	installClient.SetRegistryClient(registryClient)

	// TODO: Check if repository exists, or add it
	chartPath, err := installClient.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", r.repository, r.name), settings)
	if err != nil {
		return err
	}

	providers := getter.All(settings)

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	if chartDependencies := chart.Metadata.Dependencies; chartDependencies != nil {
		if err := action.CheckDependencies(chart, chartDependencies); err != nil {
			err = fmt.Errorf("failed to check chart dependencies: %v", err)
			if !installClient.DependencyUpdate {
				return err
			}

			manager := &downloader.Manager{
				Out:              os.Stdout,
				ChartPath:        chartPath,
				Keyring:          installClient.ChartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providers,
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
				Debug:            settings.Debug,
				RegistryClient:   installClient.GetRegistryClient(),
			}
			if err := manager.Update(); err != nil {
				return err
			}
			if chart, err = loader.Load(chartPath); err != nil {
				return err
			}
		}
	}

	_, err = installClient.RunWithContext(context.Background(), chart, r.values)
	if err != nil {
		return err
	}

	return nil
}

func newRegistryClient(settings *cli.EnvSettings, plainHTTP bool) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	}
	if plainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	// Create a new registry client
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return registryClient, nil
}

func newRegistryClientTLS(settings *cli.EnvSettings, logger *log.Logger, certFile, keyFile, caFile string, insecureSkipTLSVerify, plainHTTP bool) (*registry.Client, error) {
	if certFile != "" && keyFile != "" || caFile != "" || insecureSkipTLSVerify {
		registryClient, err := registry.NewRegistryClientWithTLS(logger.Writer(), certFile, keyFile, caFile, insecureSkipTLSVerify, settings.RegistryConfig, settings.Debug)
		if err != nil {
			return nil, err
		}

		return registryClient, nil
	}
	registryClient, err := newRegistryClient(settings, plainHTTP)
	if err != nil {
		return nil, err
	}

	return registryClient, err
}
