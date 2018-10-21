package aws

import "github.com/aws/aws-sdk-go/service/ecs"

// KeyValuePairGetValue searches the array of values and returns the matching name or nil if none are found.
func KeyValuePairGetValue(name string, values []*ecs.KeyValuePair) (*string, bool) {
	for _, item := range values {
		if *item.Name == name {
			return item.Value, true
		}
	}

	return nil, false
}
