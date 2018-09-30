package images // import "github.com/docker/docker/daemon/images"

import (
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/image"
)

// usecase 1: docker의 registry를 변경한다.
// docker tag docker.io/inkihwang/ubuntu:latest gcr.io/abcd/ubuntu:latest

// usecase 2: docker image의 tag를 변경한다.
// docker tag gcr.io/abcd/ubuntu:latest gcr.io/abcd/ubuntu:v1.1.0

// docker tag {Image ID or (Domain/Path)/Name:(Tag)} {New (Domain/Path)/Name:(New Tag)}

// TagImage creates the tag specified by newTag, pointing to the image named
// imageName (alternatively, imageName can also be an image ID).

// imageName : original Image Name or ID
// repository : new Image Name with latest?
// tag : new tag
func (i *ImageService) TagImage(imageName, repository, tag string) (string, error) {
	img, err := i.GetImage(imageName)
	if err != nil {
		return "", err
	}

	newTag, err := reference.ParseNormalizedNamed(repository)
	if err != nil {
		return "", err
	}
	if tag != "" {
		if newTag, err = reference.WithTag(reference.TrimNamed(newTag), tag); err != nil {
			return "", err
		}
	}

	err = i.TagImageWithReference(img.ID(), newTag)
	return reference.FamiliarString(newTag), err
}

// TagImageWithReference adds the given reference to the image ID provided.
// 새로만든 tag를 Image Digest에 attach
func (i *ImageService) TagImageWithReference(imageID image.ID, newTag reference.Named) error {
	if err := i.referenceStore.AddTag(newTag, imageID.Digest(), true); err != nil {
		return err
	}

	if err := i.imageStore.SetLastUpdated(imageID); err != nil {
		return err
	}
	i.LogImageEvent(imageID.String(), reference.FamiliarString(newTag), "tag")
	return nil
}
