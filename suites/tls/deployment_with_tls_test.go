package tls_test

import (
	"fmt"
	"net/http"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/suites/framework"
)

func (s *TLSSuite) Test_ConfigureTLSIssuer_OK() {
	errStr, enabled := s.checkTLSPreconditions()
	if !enabled {
		s.T().Skipf(errStr)
	}

	// Given.
	var dt = controlPlane.MustGetDeploymentTarget(ctx, s.T())
	var issuer = random.AlphaNumericString(10)

	// When & Then.
	s.setUpIssuer(issuer, dt)
	s.T().Cleanup(func() {
		s.cleanTCIssuer(issuer)
	})

	// When.
	dtResponse, httpResponse, err := controlPlane.PDS.DeploymentTargetsApi.ApiDeploymentTargetsIdGet(ctx, *dt.Id).Execute()

	// Then.
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, httpResponse.StatusCode)
	s.Require().Equal(issuer, *dtResponse.TlsIssuer)
}

func (s *TLSSuite) Test_ConfigureTLSRequired_OK() {
	errStr, enabled := s.checkTLSPreconditions()
	if !enabled {
		s.T().Skipf(errStr)
	}
	// Given.
	var dt = controlPlane.MustGetDeploymentTarget(ctx, s.T())
	var issuer = random.AlphaNumericString(10)

	s.setUpIssuer(issuer, dt)
	s.T().Cleanup(func() {
		s.cleanTCIssuer(issuer)
	})

	// When & Then.
	s.setUpRequiredTLS(issuer, dt)
	s.T().Cleanup(func() {
		s.cleanupCP(dt)
	})

	// When.
	dtResponse, httpResponse, err := controlPlane.PDS.DeploymentTargetsApi.ApiDeploymentTargetsIdGet(ctx, *dt.Id).Execute()

	// Then.
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, httpResponse.StatusCode)
	s.Require().True(*dtResponse.TlsRequired)
}

func (s *TLSSuite) Test_CreateDeploymentWithoutTLS_WhenTLSRequired_Fail() {
	errStr, enabled := s.checkTLSPreconditions()
	if !enabled {
		s.T().Skipf(errStr)
	}
	// Given.
	var dt = controlPlane.MustGetDeploymentTarget(ctx, s.T())
	var issuer = random.AlphaNumericString(10)

	s.setUpIssuer(issuer, dt)
	s.T().Cleanup(func() {
		s.cleanTCIssuer(issuer)
	})

	s.setUpRequiredTLS(issuer, dt)
	s.T().Cleanup(func() {
		s.cleanupCP(dt)
	})

	deploymentSpec := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
		TLSEnabled:      false,
	}

	// When.
	_, err := controlPlane.DeployDeploymentSpec(ctx, &deploymentSpec, controlPlane.TestPDSNamespaceID)

	// Then.
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "422")
	s.Require().Contains(err.Error(), "policy requires enabling TLS for this deployment")
}

func (s *TLSSuite) Test_CreateDeploymentWithTLS_WhenTLSRequired_OK() {
	errStr, enabled := s.checkTLSPreconditions()
	if !enabled {
		s.T().Skipf(errStr)
	}
	// Given.
	var dt = controlPlane.MustGetDeploymentTarget(ctx, s.T())
	var issuer = random.AlphaNumericString(10)

	s.setUpIssuer(issuer, dt)
	s.T().Cleanup(func() {
		s.cleanTCIssuer(issuer)
	})

	s.setUpRequiredTLS(issuer, dt)
	s.T().Cleanup(func() {
		s.cleanupCP(dt)
	})

	deploymentSpec := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
		TLSEnabled:      true,
	}

	// When.
	deploymentID, err := controlPlane.DeployDeploymentSpec(ctx, &deploymentSpec, controlPlane.TestPDSNamespaceID)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})

	// Then.
	s.Require().NoError(err)
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	controlPlane.MustWaitForDeploymentAvailable(ctx, s.T(), deploymentID)
}

func (s *TLSSuite) checkTLSPreconditions() (string, bool) {
	// Pre-condition at TC
	if !targetCluster.PDSChartConfig.DataServiceTLSEnabled {
		return "DataServiceTLSEnabled not enabled on TC", false
	}

	// Pre-condition at CP
	request := controlPlane.PDS.AccountsApi.ApiAccountsIdGet(ctx, controlPlane.TestPDSAccountID)
	account, httpResponse, err := request.Execute()
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, httpResponse.StatusCode)
	if account != nil && account.GlobalConfig != nil && account.GlobalConfig.TlsPreviewEnabled != nil {
		if *account.GlobalConfig.TlsPreviewEnabled == "all" {
			return "", true
		}
	}
	return fmt.Sprintf("TLS not configured for account %s", framework.PDSAccountName), false
}

func (s *TLSSuite) setUpIssuer(issuer string, dt *pds.ModelsDeploymentTarget) {
	err := targetCluster.CreateClusterIssuer(ctx, getClusterIssuer(issuer))
	s.Require().NoError(err)

	patchBody := pds.RequestsPatchDeploymentTargetRequest{
		TlsIssuer: &issuer,
	}
	patch := controlPlane.PDS.DeploymentTargetsApi.ApiDeploymentTargetsIdPatch(ctx, *dt.Id)
	patch = patch.Body(patchBody)
	dtResponse, httpResponse, err := patch.Execute()
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, httpResponse.StatusCode)
	s.Require().Equal(issuer, *dtResponse.TlsIssuer)
}

func getClusterIssuer(issuer string) *certmanagerv1.ClusterIssuer {
	specs := certmanagerv1.IssuerSpec{
		IssuerConfig: certmanagerv1.IssuerConfig{SelfSigned: &certmanagerv1.SelfSignedIssuer{}},
	}
	return &certmanagerv1.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterIssuer",
			APIVersion: "cert-manager.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: issuer,
		},
		Spec: specs,
	}
}

func (s *TLSSuite) cleanTCIssuer(issuer string) {
	err := targetCluster.DeleteClusterIssuer(ctx, getClusterIssuer(issuer))
	s.Require().NoError(err)
}

func (s *TLSSuite) cleanupCP(dt *pds.ModelsDeploymentTarget) {
	s.T().Cleanup(func() {
		issuer := ""
		tlsRequired := false
		patchBody := pds.RequestsPatchDeploymentTargetRequest{
			TlsRequired: &tlsRequired,
			TlsIssuer:   &issuer,
		}

		patch := controlPlane.PDS.DeploymentTargetsApi.ApiDeploymentTargetsIdPatch(ctx, *dt.Id)
		patch = patch.Body(patchBody)
		target, resp, err := patch.Execute()
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().Equal(http.StatusOK, resp.StatusCode)
		s.Require().NotNil(target)
		s.Require().NotNil(target.TlsIssuer)
		s.Require().Equal(*target.TlsIssuer, "")
		s.Require().NotNil(target.TlsRequired)
		s.Require().False(*target.TlsRequired)
	})
}

func (s *TLSSuite) setUpRequiredTLS(issuer string, dt *pds.ModelsDeploymentTarget) {
	tlsRequired := true
	patchBody := pds.RequestsPatchDeploymentTargetRequest{
		TlsRequired: &tlsRequired,
	}
	patch := controlPlane.PDS.DeploymentTargetsApi.ApiDeploymentTargetsIdPatch(ctx, *dt.Id)
	patch = patch.Body(patchBody)
	dtResponse, httpResponse, err := patch.Execute()

	s.Require().NoError(err)
	s.Require().Equal(200, httpResponse.StatusCode)
	s.Require().Equal(issuer, *dtResponse.TlsIssuer)
}
