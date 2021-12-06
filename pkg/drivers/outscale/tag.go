package outscale

import (
	"fmt"

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

	_, httpRes, err := oscApi.client.TagApi.CreateTags(oscApi.context).CreateTagsRequest(request).Execute()
	if err != nil {
		log.Error("Error while submitting the CreateTag request: ")
		if httpRes != nil {
			fmt.Printf(httpRes.Status)
		}
		return err
	}

	return nil
}
