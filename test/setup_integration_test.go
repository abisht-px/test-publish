package test

func (suite *PDSTestSuite) TestClusterSetup() {
	// If this empty test fails, the suite setup itself is broken.
	suite.T().Log("PDS test clusters successfully set up.")
}
