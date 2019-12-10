package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type executeProtecodeScanOptions struct {
	ProtecodeExcludeCVEs                 string `json:"protecodeExcludeCVEs,omitempty"`
	ProtecodeFailOnSevereVulnerabilities bool   `json:"protecodeFailOnSevereVulnerabilities,omitempty"`
	DockerRegistryURL                    string `json:"dockerRegistryUrl,omitempty"`
	CleanupMode                          string `json:"cleanupMode,omitempty"`
	FilePath                             string `json:"filePath,omitempty"`
	AddSideBarLink                       bool   `json:"addSideBarLink,omitempty"`
	Verbose                              bool   `json:"verbose,omitempty"`
	ProtecodeTimeoutMinutes              string `json:"protecodeTimeoutMinutes,omitempty"`
	ProtecodeServerURL                   string `json:"protecodeServerUrl,omitempty"`
	ReportFileName                       string `json:"reportFileName,omitempty"`
	UseCallback                          bool   `json:"useCallback,omitempty"`
	FetchURL                             string `json:"fetchUrl,omitempty"`
	ProtecodeGroup                       string `json:"protecodeGroup,omitempty"`
	ReuseExisting                        bool   `json:"reuseExisting,omitempty"`
	User                                 string `json:"user,omitempty"`     //aus dem ENV holen
	Password                             string `json:"password,omitempty"` //aus dem ENV holen
	ProtecodeCredentialsID               string `json:"protecodeCredentialsId,omitempty"`
}

var myExecuteProtecodeScanOptions executeProtecodeScanOptions
var executeProtecodeScanStepConfigJSON string

// ExecuteProtecodeScanCommand Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).
func ExecuteProtecodeScanCommand() *cobra.Command {
	metadata := executeProtecodeScanMetadata()
	var createExecuteProtecodeScanCmd = &cobra.Command{
		Use:   "executeProtecodeScan",
		Short: "Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).",
		Long: `Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).

!!! info "New: Using executeProtecodeScan for Docker images on JaaS"
    **This step now also works on "Jenkins as a Service (JaaS)"!**<br />
    For the JaaS use case where the execution happens in a Kubernetes cluster without access to a Docker daemon [skopeo](https://github.com/containers/skopeo) is now used silently in the background to save a Docker image retrieved from a registry.


!!! hint "Auditing findings (Triaging)"
    Triaging is now supported by the Protecode backend and also Piper does consider this information during the analysis of the scan results though product versions are not supported by Protecode. Therefore please make sure that the ` + "`" + `fileName` + "`" + ` you are providing does either contain a stable version or that it does not contain one at all. By ensuring that you are able to triage CVEs globally on the upload file's name without affecting any other artifacts scanned in the same Protecode group and as such triaged vulnerabilities will be considered during the next scan and will not fail the build anymore.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("executeProtecodeScan")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "executeProtecodeScan", &myExecuteProtecodeScanOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeProtecodeScan(myExecuteProtecodeScanOptions)
		},
	}

	addExecuteProtecodeScanFlags(createExecuteProtecodeScanCmd)
	return createExecuteProtecodeScanCmd
}

func addExecuteProtecodeScanFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.ProtecodeExcludeCVEs, "protecodeExcludeCVEs", "[]", "DEPRECATED: Do use triaging within the Protecode UI instead")
	cmd.Flags().BoolVar(&myExecuteProtecodeScanOptions.ProtecodeFailOnSevereVulnerabilities, "protecodeFailOnSevereVulnerabilities", true, "Whether to fail the job on severe vulnerabilties or not")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.DockerRegistryURL, "dockerRegistryUrl", os.Getenv("PIPER_dockerRegistryUrl"), "The reference to the docker registry to scan with Protecode")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.CleanupMode, "cleanupMode", "binary", "Decides which parts are removed from the Protecode backend after the scan")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.FilePath, "filePath", os.Getenv("PIPER_filePath"), "The path to the file from local workspace to scan with Protecode")
	cmd.Flags().BoolVar(&myExecuteProtecodeScanOptions.AddSideBarLink, "addSideBarLink", true, "Whether to create a side bar link pointing to the report produced by Protecode or not")
	cmd.Flags().BoolVar(&myExecuteProtecodeScanOptions.Verbose, "verbose", false, "Whether to log verbose information or not")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.ProtecodeTimeoutMinutes, "protecodeTimeoutMinutes", "60", "The timeout to wait for the scan to finish")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.ProtecodeServerURL, "protecodeServerUrl", "https://protecode.c.eu-de-2.cloud.sap", "The URL to the Protecode backend")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.ReportFileName, "reportFileName", "protecode_report.pdf", "The file name of the report to be created")
	cmd.Flags().BoolVar(&myExecuteProtecodeScanOptions.UseCallback, "useCallback", false, "Whether to the Protecode backend's callback or poll for results")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.FetchURL, "fetchUrl", os.Getenv("PIPER_fetchUrl"), "The URL to fetch the file to scan with Protecode which must be accessible via public HTTP GET request")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.ProtecodeGroup, "protecodeGroup", os.Getenv("PIPER_protecodeGroup"), "The Protecode group ID of your team")
	cmd.Flags().BoolVar(&myExecuteProtecodeScanOptions.ReuseExisting, "reuseExisting", false, "Whether to reuse an existing product instead of creating a new one")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.User, "user", os.Getenv("PIPER_user"), "user which is used for the protecode scan")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.Password, "password", os.Getenv("PIPER_password"), "password for the user")
	cmd.Flags().StringVar(&myExecuteProtecodeScanOptions.ProtecodeCredentialsID, "protecodeCredentialsId", os.Getenv("PIPER_protecodeCredentialsId"), "test")

	cmd.MarkFlagRequired("protecodeGroup")
}

// retrieve step metadata
func executeProtecodeScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "protecodeExcludeCVEs",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeFailOnSevereVulnerabilities",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "dockerRegistryUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "cleanupMode",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "filePath",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "addSideBarLink",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "verbose",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeTimeoutMinutes",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeServerUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "reportFileName",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "useCallback",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "fetchUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeGroup",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "reuseExisting",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "user",
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "password",
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeCredentialsId",
						Scope:     []string{},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
