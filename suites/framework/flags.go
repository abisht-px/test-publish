package framework

import (
	"flag"
)

const (
	DefaultCertManagerChartVersion = "v1.11.0"
	DefaultPDSNamespace            = "pds-system"
	DefaultPDSServiceAccountName   = "Default-AgentWriter"
	DefaultPDSTenantName           = "Default"
	DefaultPDSProjectName          = "Default"
	DefaultPDSAccountName          = "Portworx"
	DefaultS3Endpoint              = "s3.amazonaws.com"
	DefaultCertManagerNamespace    = "cert-manager"
	DefaultAWSRegion               = "us-west-2"
	DefaultIssuerTokenURL          = "https://apicentral.portworx.com/api"
)

var (
	// Control Plane flags.
	PDSAccountName     string
	PDSControlPlaneAPI string
	PDSTenantName      string
	PDSProjectName     string

	// Target Cluster flags.
	TargetClusterKubeconfig string
	DeploymentTargetName    string
	ServiceAccountName      string

	// Authentication flags.
	IssuerTokenURL     string
	IssuerClientID     string
	IssuerClientSecret string
	PDSUsername        string
	PDSPassword        string
	PDSAPIToken        string

	// Helm Chart flags.
	PDSHelmChartVersion     string
	CertManagerChartVersion string
	DataServiceTLSEnabled   bool

	// Backup Target flags.
	AWSAccessKey    string
	AWSS3Endpoint   string
	AWSSecretKey    string
	AWSRegion       string
	AWSS3BucketName string

	// Test Namespace
	TestNamespace string

	// Dataservice Flags
	DSVersionMatrixFile string
)

func ControlPlaneFlags() {
	flag.StringVar(&PDSAccountName, "accountName", DefaultPDSAccountName, "PDS account name")
	flag.StringVar(&PDSTenantName, "tenantName", DefaultPDSTenantName, "PDS Tenant name")
	flag.StringVar(&PDSProjectName, "projectName", DefaultPDSProjectName, "PDS Project name")
	flag.StringVar(&PDSControlPlaneAPI, "controlPlaneAPI", "", "Control Plane API Address")
}

func TargetClusterFlags() {
	flag.StringVar(
		&PDSHelmChartVersion,
		"pdsHelmChartVersion",
		"0",
		`PDS Helm Chart Version. 
\n - if value is 0 helm installation is skipped
\n - if value is empty, chart version is fetched from CP`,
	)
	flag.StringVar(
		&CertManagerChartVersion,
		"certManagerChartVersion",
		DefaultCertManagerChartVersion,
		"PDS Helm Chart Version",
	)
	flag.StringVar(
		&TargetClusterKubeconfig,
		"targetClusterKubeconfig",
		"",
		"Path to target cluster's kubeconfig. For running tests within the cluster set this as empty",
	)
	flag.StringVar(
		&DeploymentTargetName,
		"deploymentTargetName",
		"",
		"Deployment Target Name of the cluster",
	)
	flag.StringVar(
		&ServiceAccountName,
		"serviceAccountName",
		DefaultPDSServiceAccountName,
		"Service account name",
	)
	flag.StringVar(
		&TestNamespace,
		"testNamespace",
		"",
		"Test namespace to run tests",
	)

	flag.BoolVar(
		&DataServiceTLSEnabled,
		"dataServicesTLSEnabled",
		false,
		"Flag for data services TLS configuration",
	)

	if DeploymentTargetName == "" {
		DeploymentTargetName = NewRandomName("tc")
	}
}

func AuthenticationFlags() {
	flag.StringVar(&IssuerTokenURL, "issuerTokenURL", DefaultIssuerTokenURL, "Px Central URL")
	flag.StringVar(&IssuerClientID, "issuerClientID", "4", "Px Central ClientID")
	flag.StringVar(&IssuerClientSecret, "issuerClientSecret", "", "Px Central Issuer Secret")
	flag.StringVar(&PDSUsername, "pdsUserName", "", "PDS User Name")
	flag.StringVar(&PDSPassword, "pdsPassword", "", "PDS Password")
	flag.StringVar(&PDSAPIToken, "pdsToken", "", "PDS API token")
}

func BackupCredentialFlags() {
	flag.StringVar(&AWSS3Endpoint, "awsS3Endpoint", DefaultS3Endpoint, "AWS Endpoint")
	flag.StringVar(&AWSS3BucketName, "awsS3BucketName", "", "AWS S3 Bucket Name")
	flag.StringVar(&AWSRegion, "awsRegion", DefaultAWSRegion, "AWS Region")
	flag.StringVar(&AWSAccessKey, "awsAccessKey", "", "AWS Access Key")
	flag.StringVar(&AWSSecretKey, "awsSecretKey", "", "AWS Secret Key")
}

func DataserviceFlags() {
	flag.StringVar(&DSVersionMatrixFile, "dsVersionMatrixFile", "", "File path to Dataservice version matrix")
}
