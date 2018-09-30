package images // import "github.com/docker/docker/daemon/images"

import (
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/image"
)

// ErrImageDoesNotExist is error returned when no image can be found for a reference.
type ErrImageDoesNotExist struct {
	ref reference.Reference
}

func (e ErrImageDoesNotExist) Error() string {
	ref := e.ref
	// ref.instanceOf(reference.Named) in java
	// ref는 Named interface에 포함되어있다?
	if named, ok := ref.(reference.Named); ok {
		ref = reference.TagNameOnly(named)
	}
	return fmt.Sprintf("No such image: %s", reference.FamiliarString(ref))
}

// NotFound implements the NotFound interface
func (e ErrImageDoesNotExist) NotFound() {}

// 1. refOrID가 digest인가, familar name인가?
// 2. refOrID가 digetst라면 digst로 image의 ID를 얻는다.
// 3. image의 ID로 imageStore에서 이미지를 검색해서 반환한다
// 4. refOrID가 familar name이라면 name으로 digst를 referenceStore에서 검색한다.
// 5. digist로 image의 ID를 얻는다.
// 6. image의 ID로 imageStore에서 이미지를 검색해서 반환한다.
// 중간에 에러가 발생하면 ErrImageDoesNotExist 리턴

// GetImage returns an image corresponding to the image referred to by refOrID.
func (i *ImageService) GetImage(refOrID string) (*image.Image, error) {
	ref, err := reference.ParseAnyReference(refOrID)
	if err != nil {
		return nil, errdefs.InvalidParameter(err)
	}
	namedRef, ok := ref.(reference.Named)
	if !ok {
		digested, ok := ref.(reference.Digested)
		if !ok {
			return nil, ErrImageDoesNotExist{ref}
		}
		id := image.IDFromDigest(digested.Digest())
		// 왜 err != nil 로 안하는지?
		if img, err := i.imageStore.Get(id); err == nil {
			return img, nil
		}
		return nil, ErrImageDoesNotExist{ref}
	}

	if digest, err := i.referenceStore.Get(namedRef); err == nil {
		// Search the image stores to get the operating system, defaulting to host OS.
		id := image.IDFromDigest(digest)
		if img, err := i.imageStore.Get(id); err == nil {
			return img, nil
		}
	}

	// Search based on ID
	if id, err := i.imageStore.Search(refOrID); err == nil {
		img, err := i.imageStore.Get(id)
		if err != nil {
			return nil, ErrImageDoesNotExist{ref}
		}
		return img, nil
	}

	return nil, ErrImageDoesNotExist{ref}
}
