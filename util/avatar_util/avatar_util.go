package avatarutil

import (
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

func ProcessAvatar(source string) (result *string, err error) {
	buffer, err := bimg.Read("./file/image/" + source)
	if err != nil {
		return nil, err
	}

	width := 250
	height := 250
	newImage, err := bimg.NewImage(buffer).SmartCrop(width, height)

	size, _ := bimg.Size(newImage)
	if size.Width != width || size.Height != height {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Image"}
	}

	newID := uuid.New()
	resultfilename := newID.String() + ".webp"

	err = bimg.Write("./file/image/"+resultfilename, newImage)
	if err != nil {
		return nil, err
	}

	return &resultfilename, nil
}
