package avatarutil

import (
	"github.com/disintegration/imaging"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
)

func ProcessAvatar(source string) (result *string, err error) {
	srcImage, err := imaging.Open("./file/image/" + source)
	if err != nil {
		return nil, err
	}

	width := 250
	height := 250
	dstImage := imaging.Fill(srcImage, width, height, imaging.Center, imaging.Lanczos)

	size := dstImage.Rect.Size()
	if size.X != width || size.Y != height {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Image"}
	}

	newID := uuid.New()
	resultfilename := newID.String() + ".jpg"

	err = imaging.Save(dstImage, "./file/image/"+resultfilename)
	if err != nil {
		return nil, err
	}

	return &resultfilename, nil
}
