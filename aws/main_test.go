package aws

import (
	"os"
	"testing"
)

func PreTest(t *testing.T) {
	if len(os.Getenv("ACC")) == 0 {
		t.Skip("ACC environment variable needs to be declared to run this test")
	}
}
