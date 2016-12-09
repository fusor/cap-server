package main

import (
	"fmt"
	"github.com/containers/image/transports"
	"github.com/containers/image/types"
	"time"
)

type inspectOutput struct {
	Name          string `json:",omitempty"`
	Tag           string `json:",omitempty"`
	Digest        string
	RepoTags      []string
	Created       time.Time
	DockerVersion string
	Labels        map[string]string
	Architecture  string
	Os            string
	Layers        []string
}

func IsImageAtomicApp(image_name string, channel chan<- string) {
	img, err := parseImage(image_name)
	if err != nil {
		fmt.Println(err)
		channel <- fmt.Sprintf("ignore\n")
		return
	}
	defer img.Close()

	imgInspect, err := img.Inspect()
	if err != nil {
		fmt.Println(err)
		channel <- fmt.Sprintf("ignore\n")
		return
	}

	outputData := inspectOutput{
		Name: "", // Possibly overridden for a docker.Image.
		Tag:  imgInspect.Tag,
		// Digest is set below.
		RepoTags:      []string{}, // Possibly overriden for a docker.Image.
		Created:       imgInspect.Created,
		DockerVersion: imgInspect.DockerVersion,
		Labels:        imgInspect.Labels,
		Architecture:  imgInspect.Architecture,
		Os:            imgInspect.Os,
		Layers:        imgInspect.Layers,
	}

	if outputData.Labels["io.projectatomic.nulecule.atomicappversion"] != "" {
		channel <- fmt.Sprintf("%s\n", image_name)
	} else {
		channel <- fmt.Sprintf("ignore\n")
	}
}

// ParseImage converts image URL-like string to an initialized handler for that image.
// The caller must call .Close() on the returned Image.
func parseImage(name string) (types.Image, error) {
	imgName := name
	ref, err := transports.ParseImageName(imgName)
	if err != nil {
		return nil, err
	}
	return ref.NewImage(contextFromGlobalOptions(false))
}

// contextFromGlobalOptions returns a types.SystemContext depending on c.
func contextFromGlobalOptions(tls bool) *types.SystemContext {
	tlsVerify := tls
	return &types.SystemContext{
		RegistriesDirPath:           "",
		DockerCertPath:              "",
		DockerInsecureSkipTLSVerify: !tlsVerify,
	}
}
