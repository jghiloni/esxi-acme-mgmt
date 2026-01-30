package common

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"plugin"
	"strings"
)

const SymbolName = "DNSProvider"

func LoadProvider(pluginDir string, providerName string, providerArgs []string) (Provider, error) {
	var provider Provider
	p, err := plugin.Open(filepath.Join(pluginDir, providerName+".so"))
	if err != nil {
		provider, err = findProvider(pluginDir, providerName)
		if err != nil {
			return nil, fmt.Errorf("could not provider %s: %w", providerName, err)
		}
		provider.WithArgs(providerArgs)
		return provider, nil
	}

	provider, err = loadProviderFromPlugin(p, providerName)
	if provider != nil {
		provider.WithArgs(providerArgs)
	}
	return provider, err
}

func findProvider(pluginDir, providerName string) (Provider, error) {
	logger := slog.With(slog.String("method", "findProvider"), slog.String("provider-name", providerName))
	pluginFiles, _ := filepath.Glob(filepath.Join(pluginDir, "*.so"))

	logger.Debug("found plugin files", slog.Any("plugin-files", pluginFiles))
	var errs []error
	for _, soFile := range pluginFiles {
		p, e := plugin.Open(soFile)
		errs = append(errs, e)
		if p != nil {
			provider, err := loadProviderFromPlugin(p, providerName)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			return provider, nil
		}
	}

	return nil, errors.Join(errs...)
}

func loadProviderFromPlugin(p *plugin.Plugin, providerName string) (Provider, error) {
	raw, err := p.Lookup(SymbolName)
	if err != nil {
		return nil, err
	}

	if provider, asserted := raw.(Provider); asserted {
		if strings.EqualFold(provider.Name(), providerName) {
			return provider, nil
		}

		return nil, fmt.Errorf("expected provider name %s, found %s", providerName, provider.Name())
	}

	return nil, fmt.Errorf("exported symbol %s is not a Provider, it is a %T", SymbolName, raw)
}
