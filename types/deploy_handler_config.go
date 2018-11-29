package types

// DeployHandlerConfig specify options for Deployments
type DeployHandlerConfig struct {
	AssignPublicIP  string
	SecurityGroupID string
	SubnetIDs       string
	VpcID           string
	Region          string
}
