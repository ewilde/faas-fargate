package function

import (
	"fmt"
	"io/ioutil"
)

// Handle a serverless request
func Handle(req []byte) string {
	dat, err := ioutil.ReadFile("/run/secrets/openfaas-db-password")
	if err != nil {
		return fmt.Sprintf("Hello, Go. Error: %v", err)
	}

	return fmt.Sprintf("Hello, Go. You said: %s, my secret is: %s", string(req), string(dat))
}
