package types

// DeployHandlerConfig specify options for Deployments
type DeployHandlerConfig struct {
	AssignPublicIP  string
	SecurityGroupId string
	SubnetIDs       string
	VpcID           string
}
