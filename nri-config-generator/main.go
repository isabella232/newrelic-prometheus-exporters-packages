package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/log"
	"github.com/newrelic/nri-config-generator/args"
	"github.com/newrelic/nri-config-generator/generator"
	"github.com/newrelic/nri-config-generator/httport"
)

const (
	varIntegrationName    = "integration"
	varExporterPort       = "exporter_port"
	varExporterDefinition = "exporter_definition"
	sleepTime             = 30 * time.Second
)

var (
	integration string
	//go:embed templates
	integrationTemplate embed.FS
	//go:embed templates/config.json.tmpl
	configTemplateContent string
	vars                  = map[string]interface{}{
		varIntegrationName: integration,
	}
)

func main() {
	err := args.PopulateVars(vars)
	panicErr(err)
	integrationTemplatePattern := fmt.Sprintf("templates/%s.json.tmpl", integration)

	content, err := integrationTemplate.ReadFile(integrationTemplatePattern)
	panicErr(err)

	integrationTemplate, err := loadIntegrationTemplate(content)
	panicErr(err)
	exporterGenerator := generator.NewExporter(integration, integrationTemplate)
	configTemplate, err := loadConfigTemplate()
	panicErr(err)
	configGenerator := generator.NewConfig(configTemplate)
	port, err := findExporterPort()
	panicErr(err)
	vars[varExporterPort] = fmt.Sprintf("%v",port)
	exporterText, err := exporterGenerator.Generate(vars)
	panicErr(err)
	vars[varExporterDefinition] = exporterText
	output, err := configGenerator.Generate(vars)
	panicErr(err)
	fmt.Println(output)
	httport.SetPrometheusExporterPort("localhost", port)
	for {
		time.Sleep(sleepTime)

		if !httport.IsPrometheusExporterRunning() {
			panicErr(errors.New("there is not a prometheus exporter in the assigned port"))
		}
		fmt.Println("{}")
		fmt.Println(output)
	}

}

func panicErr(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func findExporterPort() (int, error) {
	cfgPort := ""
	if cfg, ok := vars[args.PrefixCfg]; ok {
		cfgVars := cfg.(map[string]interface{})
		if cfgPortPtr := cfgVars[varExporterPort]; cfgPortPtr != nil {
			cfgPort = fmt.Sprintf("%v", cfgPortPtr)
		}
	}
	port, err := httport.GetAvailablePort(cfgPort)
	if err != nil {
		log.Error("error obtaining the port for the exporter: '%s'", err.Error())
		return -1, err
	}
	return port, nil
}

func loadIntegrationTemplate(content []byte) (*template.Template, error) {
	t, err := template.New("").Funcs(generator.TemplatesFunc).Parse(string(content))
	if err != nil {
		log.Error("error parsing the integration template: '%s'", err.Error())
		return nil, err
	}
	return t, nil
}

func loadConfigTemplate() (*template.Template, error) {
	t, err := template.New("").Funcs(generator.TemplatesFunc).Parse(configTemplateContent)
	if err != nil {
		log.Error("error parsing the template for the config: '%s'", err.Error())
		return nil, err
	}
	return t, nil
}
