package namespace_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/portworx/pds-integration-test/internal/random"
)

// TestNamespaceSync_OK tests the successful creation and status of the namespace.
// Steps:
// 1. Add a randomised namespace name
// 2. Assign Metadata with a proper label
// 3. Create a namespace
// 4. Check for the namespace status
// Expected:
// 1. Creation of namespace should be success
// 2. Namespace status should be available
func (s *NamespaceTestSuite) TestNamespaceSync_OK() {
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
	s.controlPlane.MustWaitForNamespaceStatus(s.ctx, s.T(), namespaceName, "available")
}

// TestNamespaceSync_MissingLabel tests the not found status of the namespace.
// Steps:
// 1. Add a randomised namespace name
// 2. Assign Metadata with wrong/empty label
// 3. Create a namespace with these values
// 4. Check for the namespace status
// Expected:
// 1. Creation of namespace should not happen
// 2. Namespace status - not expected to be found in control plane
func (s *NamespaceTestSuite) TestNamespaceSync_MissingLabel() {
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
		s.T().Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Given.
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "integration-test-" + random.AlphaNumericString(10),
					Labels: c.labels,
				},
			}

			// When.
			_, err := s.targetCluster.CreateNamespace(s.ctx, namespace)
			t.Cleanup(func() { _ = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name) })
			require.NoError(t, err)

			// Then.
			s.controlPlane.MustNeverGetNamespaceByName(s.ctx, t, namespace.Name)
		})
	}
}

// TestNamespaceSync_TerminatingNamespace tests the terminating status of the namespace.
// Steps:
// 1. Add a randomised namespace name
// 2. Assign Metadata with a proper label and a Finalizer
// 3. Create a namespace with these values
// 4. Check for the namespace status
// 5. Put namespace into Terminating state
// 6. Check for the namespace status again
// Expected:
// 1. Creation of namespace should be success
// 2. Namespace status should be available
// 3. Namespace is deleted
// 4. Namespace status should change to unavailable
func (s *NamespaceTestSuite) TestNamespaceSync_TerminatingNamespace() {
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

	s.controlPlane.MustWaitForNamespaceStatus(s.ctx, s.T(), namespace.Name, "available")

	// When.
	// Turn the namespace into 'Terminating' state because of the finalizer.
	err = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name)
	s.Require().NoError(err)

	// Then.
	s.controlPlane.MustWaitForNamespaceStatus(s.ctx, s.T(), namespace.Name, "unavailable")
}

// TestNamespaceSync_DeleteLabel tests the label association with status of the namespace.
// Steps:
// 1. Add a randomised namespace name
// 2. Assign Metadata with a proper label
// 3. Create a namespace with these values
// 4. Check for the namespace status
// 5. Set the labels as nil
// 6. Check for the namespace status again
// Expected:
// 1. Creation of namespace should be success
// 2. Namespace status should be available
// 3. Label is deleted
// 4. Namespace status should change to unavailable
func (s *NamespaceTestSuite) TestNamespaceSync_DeleteLabel() {
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
	s.controlPlane.MustWaitForNamespaceStatus(s.ctx, s.T(), namespace.Name, "available")

	// When.
	createdNamespace.SetLabels(nil)
	_, err = s.targetCluster.UpdateNamespace(s.ctx, createdNamespace)
	s.Require().NoError(err)

	// Then.
	s.controlPlane.MustWaitForNamespaceStatus(s.ctx, s.T(), namespace.Name, "unavailable")
}

// TestNamespaceSync_DeleteNamespace tests the deletion of namespace.
// Steps:
// 1. Add a randomised namespace name
// 2. Assign Metadata with a proper label
// 3. Create a namespace with these values
// 4. Check for the namespace status
// 5. Delete the namespace
// 6. Check for the namespace status again
// Expected:
// 1. Creation of namespace should be success
// 2. Namespace status should be available
// 3. Namespace is deleted
// 4. Namespace status should change to unavailable
func (s *NamespaceTestSuite) TestNamespaceSync_DeleteNamespace() {
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
	s.controlPlane.MustWaitForNamespaceStatus(s.ctx, s.T(), namespace.Name, "available")

	// When.
	err = s.targetCluster.DeleteNamespace(s.ctx, namespace.Name)
	s.Require().NoError(err)

	// Then.
	s.controlPlane.MustWaitForNamespaceStatus(s.ctx, s.T(), namespace.Name, "unavailable")
}
