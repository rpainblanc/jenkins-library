package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type protecodeExecuteScanOptions struct {
	ExcludeCVEs                 string `json:"excludeCVEs,omitempty"`
	FailOnSevereVulnerabilities bool   `json:"failOnSevereVulnerabilities,omitempty"`
	ScanImage                   string `json:"scanImage,omitempty"`
	DockerRegistryURL           string `json:"dockerRegistryUrl,omitempty"`
	CleanupMode                 string `json:"cleanupMode,omitempty"`
	FilePath                    string `json:"filePath,omitempty"`
	IncludeLayers               bool   `json:"includeLayers,omitempty"`
	AddSideBarLink              bool   `json:"addSideBarLink,omitempty"`
	TimeoutMinutes              string `json:"timeoutMinutes,omitempty"`
	ServerURL                   string `json:"serverUrl,omitempty"`
	ReportFileName              string `json:"reportFileName,omitempty"`
	FetchURL                    string `json:"fetchUrl,omitempty"`
	Group                       string `json:"group,omitempty"`
	ReuseExisting               bool   `json:"reuseExisting,omitempty"`
	User                        string `json:"user,omitempty"`
	Password                    string `json:"password,omitempty"`
	ArtifactVersion             string `json:"artifactVersion,omitempty"`
	PullRequestName             string `json:"pullRequestName,omitempty"`
}

type protecodeExecuteScanInflux struct {
	protecode_data struct {
		fields struct {
			historical_vulnerabilities string
			triaged_vulnerabilities    string
			excluded_vulnerabilities   string
			minor_vulnerabilities      string
			major_vulnerabilities      string
			vulnerabilities            string
		}
		tags struct {
		}
	}
}

func (i *protecodeExecuteScanInflux) persist(path, resourceName string) {
	measurementContent := []struct {
		measurement string
		valType     string
		name        string
		value       string
	}{
		{valType: config.InfluxField, measurement: "protecode_data", name: "historical_vulnerabilities", value: i.protecode_data.fields.historical_vulnerabilities},
		{valType: config.InfluxField, measurement: "protecode_data", name: "triaged_vulnerabilities", value: i.protecode_data.fields.triaged_vulnerabilities},
		{valType: config.InfluxField, measurement: "protecode_data", name: "excluded_vulnerabilities", value: i.protecode_data.fields.excluded_vulnerabilities},
		{valType: config.InfluxField, measurement: "protecode_data", name: "minor_vulnerabilities", value: i.protecode_data.fields.minor_vulnerabilities},
		{valType: config.InfluxField, measurement: "protecode_data", name: "major_vulnerabilities", value: i.protecode_data.fields.major_vulnerabilities},
		{valType: config.InfluxField, measurement: "protecode_data", name: "vulnerabilities", value: i.protecode_data.fields.vulnerabilities},
	}

	errCount := 0
	for _, metric := range measurementContent {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(metric.measurement, fmt.Sprintf("%vs", metric.valType), metric.name), metric.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting influx environment.")
			errCount++
		}
	}
	if errCount > 0 {
		os.Exit(1)
	}
}

// ProtecodeExecuteScanCommand Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).
func ProtecodeExecuteScanCommand() *cobra.Command {
	metadata := protecodeExecuteScanMetadata()
	var stepConfig protecodeExecuteScanOptions
	var startTime time.Time
	var influx protecodeExecuteScanInflux

	var createProtecodeExecuteScanCmd = &cobra.Command{
		Use:   "protecodeExecuteScan",
		Short: "Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).",
		Long: `Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).

!!! info "New: Using protecodeExecuteScan for Docker images on JaaS"
    **This step now also works on "Jenkins as a Service (JaaS)"!**<br />
    For the JaaS use case where the execution happens in a Kubernetes cluster without access to a Docker daemon [skopeo](https://github.com/containers/skopeo) is now used silently in the background to save a Docker image retrieved from a registry.


!!! hint "Auditing findings (Triaging)"
    Triaging is now supported by the Protecode backend and also Piper does consider this information during the analysis of the scan results though product versions are not supported by Protecode. Therefore please make sure that the ` + "`" + `fileName` + "`" + ` you are providing does either contain a stable version or that it does not contain one at all. By ensuring that you are able to triage CVEs globally on the upload file's name without affecting any other artifacts scanned in the same Protecode group and as such triaged vulnerabilities will be considered during the next scan and will not fail the build anymore.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("protecodeExecuteScan")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "protecodeExecuteScan", &stepConfig, config.OpenPiperFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				influx.persist(GeneralConfig.EnvRootPath, "influx")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, "protecodeExecuteScan")
			protecodeExecuteScan(stepConfig, &telemetryData, &influx)
			telemetryData.ErrorCode = "0"
		},
	}

	addProtecodeExecuteScanFlags(createProtecodeExecuteScanCmd, &stepConfig)
	return createProtecodeExecuteScanCmd
}

func addProtecodeExecuteScanFlags(cmd *cobra.Command, stepConfig *protecodeExecuteScanOptions) {
	cmd.Flags().StringVar(&stepConfig.ExcludeCVEs, "excludeCVEs", "[]", "DEPRECATED: Do use triaging within the Protecode UI instead")
	cmd.Flags().BoolVar(&stepConfig.FailOnSevereVulnerabilities, "failOnSevereVulnerabilities", true, "Whether to fail the job on severe vulnerabilties or not")
	cmd.Flags().StringVar(&stepConfig.ScanImage, "scanImage", os.Getenv("PIPER_scanImage"), "The reference to the docker image to scan with Protecode")
	cmd.Flags().StringVar(&stepConfig.DockerRegistryURL, "dockerRegistryUrl", os.Getenv("PIPER_dockerRegistryUrl"), "The reference to the docker registry to scan with Protecode")
	cmd.Flags().StringVar(&stepConfig.CleanupMode, "cleanupMode", "binary", "Decides which parts are removed from the Protecode backend after the scan")
	cmd.Flags().StringVar(&stepConfig.FilePath, "filePath", os.Getenv("PIPER_filePath"), "The path to the file from local workspace to scan with Protecode")
	cmd.Flags().BoolVar(&stepConfig.IncludeLayers, "includeLayers", false, "Flag if the docker layers should be included")
	cmd.Flags().BoolVar(&stepConfig.AddSideBarLink, "addSideBarLink", true, "Whether to create a side bar link pointing to the report produced by Protecode or not")
	cmd.Flags().StringVar(&stepConfig.TimeoutMinutes, "timeoutMinutes", "60", "The timeout to wait for the scan to finish")
	cmd.Flags().StringVar(&stepConfig.ServerURL, "serverUrl", os.Getenv("PIPER_serverUrl"), "The URL to the Protecode backend")
	cmd.Flags().StringVar(&stepConfig.ReportFileName, "reportFileName", "protecode_report.pdf", "The file name of the report to be created")
	cmd.Flags().StringVar(&stepConfig.FetchURL, "fetchUrl", os.Getenv("PIPER_fetchUrl"), "The URL to fetch the file to scan with Protecode which must be accessible via public HTTP GET request")
	cmd.Flags().StringVar(&stepConfig.Group, "group", os.Getenv("PIPER_group"), "The Protecode group ID of your team")
	cmd.Flags().BoolVar(&stepConfig.ReuseExisting, "reuseExisting", false, "Whether to reuse an existing product instead of creating a new one")
	cmd.Flags().StringVar(&stepConfig.User, "user", os.Getenv("PIPER_user"), "User which is used for the protecode scan")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password which is used for the user")
	cmd.Flags().StringVar(&stepConfig.ArtifactVersion, "artifactVersion", os.Getenv("PIPER_artifactVersion"), "The version of the artifact to allow identification in protecode backend")
	cmd.Flags().StringVar(&stepConfig.PullRequestName, "pullRequestName", os.Getenv("PIPER_pullRequestName"), "The name of the pull request")

	cmd.MarkFlagRequired("serverUrl")
	cmd.MarkFlagRequired("group")
	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("password")
}

// retrieve step metadata
func protecodeExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "excludeCVEs",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "protecodeExcludeCVEs"}},
					},
					{
						Name:        "failOnSevereVulnerabilities",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "protecodeFailOnSevereVulnerabilities"}},
					},
					{
						Name:        "scanImage",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "container/imageNameTag"}},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerImage"}},
					},
					{
						Name:        "dockerRegistryUrl",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "container/registryUrl"}},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "cleanupMode",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "filePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "includeLayers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "addSideBarLink",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "timeoutMinutes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "protecodeTimeoutMinutes"}},
					},
					{
						Name:        "serverUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "protecodeServerUrl"}},
					},
					{
						Name:        "reportFileName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "fetchUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "group",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "protecodeGroup"}},
					},
					{
						Name:        "reuseExisting",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "user",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "artifactVersion",
						ResourceRef: []config.ResourceReference{{Name: "commonPipelineEnvironment", Param: "artifactVersion"}},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "pullRequestName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
