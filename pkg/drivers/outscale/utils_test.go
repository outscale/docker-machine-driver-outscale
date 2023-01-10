package outscale

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThrottling(t *testing.T) {
	tags := map[error]bool{
		nil:            false,
		fmt.Errorf("not an cloudError"):       false,
		CloudError{httpRes: &http.Response{StatusCode: 500}}:       false,
		CloudError{httpRes: nil, cloudError: fmt.Errorf("Hello Error")}:       false,
		CloudError{httpRes: &http.Response{StatusCode: ThrottlingErrors[0]}}:       true,
		CloudError{httpRes: &http.Response{StatusCode: ThrottlingErrors[1]}}:       true,
	}
	for err, expected := range tags {
		res := isThrottlingError(err)

		assert.Equalf(t, expected, res, "The result is not the one expected for the error '%v'", err)
	}
}
