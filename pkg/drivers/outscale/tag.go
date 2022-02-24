package outscale

import (
	"fmt"
	"net/http"
	"strings"

	retry "github.com/avast/retry-go"
	"github.com/docker/machine/libmachine/log"
	osc "github.com/outscale/osc-sdk-go/v2"
)

func addTag(d *OscDriver, resourceId string, key string, value string) error {
	log.Debugf("Add tag {\"%s\": \"%s\"} to %s", key, value, resourceId)

	// Get the client
	oscApi, err := d.getClient()
	if err != nil {
		return err
	}

	request := osc.CreateTagsRequest{
		ResourceIds: []string{
			resourceId,
		},
		Tags: []osc.ResourceTag{
			{
				Key:   key,
				Value: value,
			},
		},
	}

	var httpRes *http.Response
	err = retry.Do(
		func() error {
			var response_error error
			_, httpRes, response_error = oscApi.client.TagApi.CreateTags(oscApi.context).CreateTagsRequest(request).Execute()
			return response_error
		},
		defaultThrottlingRetryOption...,
	)

	if err != nil {
		log.Error("Error while submitting the CreateTag request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	return nil
}

func addExtraTags(d *OscDriver, resourceId string, tags []string) error {
	if tags == nil {
		log.Debug("Skipping because there is no tags to add")
		return nil
	}

	for _, tag := range tags {
		splittedTag := strings.Split(tag, "=")
		if len(splittedTag) != 2 {
			return fmt.Errorf("The tags '%v' does not have the right syntax 'key=value'", tag)
		}
		key := splittedTag[0]
		value := splittedTag[1]

		log.Debugf("Adding tag '%v' to the resource %v", tag, resourceId)
		if err := addTag(d, resourceId, key, value); err != nil {
			return err
		}
	}
	return nil
}

func validateExtraTagsFormat(tags []string) bool {
	for _, tag := range tags {
		splittedTag := strings.Split(tag, "=")
		if len(splittedTag) != 2 {
			return false
		}
	}
	return true
}
