package test

import (
	_ "github.com/stretchr/testify/suite" // Pin import to recognize suite tests
)

func (s *PDSTestSuite) TestClusterSetup() {
	// If this empty test fails, the suite setup itself is broken.
	s.T().Log("PDS test clusters successfully set up.")
}
