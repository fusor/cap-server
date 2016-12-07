package main
import(
  "fmt"
  "github.com/containers/image/transports"
  "github.com/containers/image/types"
//  "github.com/containers/image/manifest"
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

func IsImageAtomicApp(image_name string, channel chan<- string){
  img, err := parseImage(image_name)
  if err != nil {
    fmt.Println(err)
    channel <- fmt.Sprintf("ignore\n")
    return
  }
  defer img.Close()
//  rawManifest, _, err := img.Manifest()
//  if err != nil {
//    return err
//  }
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
//  digest, err := manifest.Digest(rawManifest)
//  if err != nil {
//    return fmt.Errorf("Error computing manifest digest: %v", err)
//  }
  if outputData.Labels["io.projectatomic.nulecule.atomicappversion"]  != "" {
    fmt.Println("This repo is an atomicapp!")
    channel <- fmt.Sprintf("%s\n", image_name)
  } else {
    fmt.Println("This repo is not an atomicapp!")
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
