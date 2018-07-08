package aws

import "github.com/aws/aws-sdk-go/service/ecs"

func KeyValuePairGetValue(name string, values []*ecs.KeyValuePair) (*string, bool) {
	for _, item := range values {
		if *item.Name == name {
			return item.Value, true
		}
	}

	return nil, false
}
