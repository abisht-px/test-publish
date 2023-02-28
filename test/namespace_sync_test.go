package test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/portworx/pds-integration-test/internal/random"
)

func (s *PDSTestSuite) TestNamespaceSync_OK() {
	s.T().Parallel()
	// Given.
	namespaceName := "integration-test-" + random.AlphaNumericString(10)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				"pds.portworx.com/available": "true",
			},
		},
	}

	// When.
	_, err := s.targetCluster.CreateNamespace(s.ctx, namespace)
	s.T().Cleanup(func() { _ = s.targetCluster.DeleteNamespace(s.ctx, namespaceName) })
	s.Require().NoError(err)

	// Then.
	s.mustEventuallyGetNamespaceByName(namespaceName, "available")
}

func (s *PDSTestSuite) TestNamespaceSync_MissingLabel() {
	testcases := []struct {
		name   string
		labels map[string]string
	}{
		{
			name:   "missing label",
			labels: nil,
		},
		{
			name: "empty label value",
			labels: map[string]string{
				"pds.portworx.com/available": "",
			},
		},
		{
			name: "wrong label value",
			labels: map[string]string{
				"pds.portworx.com/available": "xxx",
			},
		},
	}

	for _, testcase := range testcases {
		// Copy testcase to make sure it runs properly in parallel.
		c := testcase
		s.Run(c.name, func() {
			s.T().Parallel()

			// Given.
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "integration-test-" + random.AlphaNumericString(10),
					Labels: c.labels,
				},
			}

			// When.
			_, err := s.targetCluster.CreateNamespace(s.ctx, namespace)
			s.T().Cleanup(func() { _ = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name) })
			s.Require().NoError(err)

			// Then.
			s.mustNeverGetNamespaceByName(namespace.Name)
		})
	}

}

func (s *PDSTestSuite) TestNamespaceSync_TerminatingNamespace() {
	s.T().Parallel()
	// Given.
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "integration-test-" + random.AlphaNumericString(10),
			Labels: map[string]string{
				"pds.portworx.com/available": "true",
			},
			Finalizers: []string{
				"pds.portworx.io/integration-test-finalizer",
			},
		},
	}

	_, err := s.targetCluster.CreateNamespace(s.ctx, namespace)
	s.T().Cleanup(func() { _ = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name) })
	s.T().Cleanup(func() { _, _ = s.targetCluster.RemoveNamespaceFinalizers(s.ctx, namespace.Name) })
	s.Require().NoError(err)

	s.mustEventuallyGetNamespaceByName(namespace.Name, "available")

	// When.
	// Turn the namespace into 'Terminating' state because of the finalizer.
	err = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name)
	s.Require().NoError(err)

	// Then.
	s.mustEventuallyGetNamespaceByName(namespace.Name, "unavailable")
}

func (s *PDSTestSuite) TestNamespaceSync_DeleteLabel() {
	s.T().Parallel()
	// Given.
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "integration-test-" + random.AlphaNumericString(10),
			Labels: map[string]string{
				"pds.portworx.com/available": "true",
			},
		},
	}

	createdNamespace, err := s.targetCluster.CreateNamespace(s.ctx, namespace)
	s.T().Cleanup(func() { _ = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name) })
	s.Require().NoError(err)
	s.mustEventuallyGetNamespaceByName(namespace.Name, "available")

	// When.
	createdNamespace.SetLabels(nil)
	_, err = s.targetCluster.UpdateNamespace(s.ctx, createdNamespace)
	s.Require().NoError(err)

	// Then.
	s.mustEventuallyGetNamespaceByName(namespace.Name, "unavailable")
}

func (s *PDSTestSuite) TestNamespaceSync_DeleteNamespace() {
	s.T().Parallel()
	// Given.
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "integration-test-" + random.AlphaNumericString(10),
			Labels: map[string]string{
				"pds.portworx.com/available": "true",
			},
		},
	}

	_, err := s.targetCluster.CreateNamespace(s.ctx, namespace)
	s.T().Cleanup(func() { _ = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name) })
	s.Require().NoError(err)
	s.mustEventuallyGetNamespaceByName(namespace.Name, "available")

	// When.
	err = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name)
	s.Require().NoError(err)

	// Then.
	s.mustEventuallyGetNamespaceByName(namespace.Name, "unavailable")
}
