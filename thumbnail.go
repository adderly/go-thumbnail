// Package thumbnail provides a method to create thumbnails from images.
package thumbnail

import (
	"errors"
	"fmt"
	"image"
	"log"
	"path/filepath"

	"github.com/sunshineplan/imgconv"
)

// An Image is an image and information about it.
type Image struct {
	// Path is a path to an image.
	Path string

	// Data is the image data in a byte-array
	ImageData image.Image

	// Current stores the existing image's dimensions
	Size ImageSize

	// Future store the new thumbnail dimensions.
	//TODO: compatibility reasons
	TargetDimension ImageSize
}
type ImageSize struct {
	Width  int
	Height int
}

// ImageDimension stores dimensional information for an Image.
type ImageDimension struct {
	// Width is the width of an image in pixels.
	Width int

	// Height is the height on an image in pixels.
	Height int

	// Percentage
	Percentage float64

	//For selecting the images there is need for the selection of the names.
	// Prefix > Name > Default [ the order of the selection of the namings]
	//Prefix
	Prefix string

	//Name
	Name string
}

type GenerationResult struct {
	// Filename the name of the file
	Filename string
	// Path the path of the file in the file system
	Path string
	//Error the error reported by the process of the generation
	Error error
}

var (
	// ErrInvalidMimeType is returned when a non-image content type is
	// detected.
	ErrInvalidMimeType            = errors.New("invalid mimetype")
	ErrInvalidImageData           = errors.New("invalid image data ")
	ErrInvalidNoTransformProvided = errors.New("no transform data was provided ")

	// ErrInvalidScaler is returned when an unrecognized scaler is
	// passed to the Generator.
	ErrInvalidScaler = errors.New("invalid scaler")

	// DefaultThumbnailPercentage the default value to use on percentage resizing
	DefaultThumbnailPercentage = 0.4

	// DefaultThumbnailSize the default dimensions used for resizing
	DefaultThumbnailSize = ImageSize{Width: 220, Height: 220}
)

// NewGenerator returns an instance of a thumbnail generator with a
// given configuration.
func NewGenerator(c Generator, outputFormats []ImageDimension) *Generator {
	return &Generator{
		Width:             300,
		Height:            300,
		DestinationPath:   c.DestinationPath,
		DestinationPrefix: c.DestinationPrefix,
		PreferredFormat:   imgconv.FormatOption{Format: imgconv.JPEG},
		OutputFormats:     outputFormats,
	}
}

// NewImageFromFile reads in an image file from the file system and
// populates an Image object. That new Image object is returned along
// with any errors that occur during the operation.
func (gen *Generator) NewImageFromFile(path string) (*Image, error) {
	// Open a test image.
	// This should not crash the program
	src, err := imgconv.Open(path)
	if err != nil {
		log.Printf("failed to open image: %v", err)
		return nil, err
	}

	return &Image{
		Path:      path,
		ImageData: src,

		Size: ImageSize{
			Width:  src.Bounds().Max.X,
			Height: src.Bounds().Max.Y,
		},
		TargetDimension: ImageSize{
			Width:  gen.Width,
			Height: gen.Height,
		},
	}, nil
}

// NewImageFromByteArray reads in an image from a byte array and
// populates an Image object. That new Image object is returned along
// with any errors that occur during the operation.
//func (gen *Generator) NewImageFromByteArray(imageBytes []byte) (*Image, error) {
//
//	return &Image{
//		Data: imageBytes,
//		Size: len(imageBytes),
//		Current: ImageDimension{
//			Width:  0,
//			Height: 0,
//		},
//		Future: ImageDimension{
//			Width:  gen.Width,
//			Height: gen.Height,
//		},
//	}, nil
//}

// Generator registers a generator configuration to be used when
// creating thumbnails.
type Generator struct {
	// Width is the destination thumbnail width.
	Width int

	// Height is the destination thumbnail height.
	Height int

	// The preferred format for exporting the thumbnails
	PreferredFormat imgconv.FormatOption

	// DestinationPath is the destination thumbnail path.
	DestinationPath string

	// DestinationPrefix is the prefix for the destination thumbnail
	// filename.
	DestinationPrefix string

	// OutputFormats the formats (dimensions), that the image will be exported to.
	OutputFormats []ImageDimension
}

// Generate an image, Profile [profile-xl , profile-sm, profile-ico]
// Source: Profile.png
// Generated Result: [ profile-xl.jpg ,  profile-sm.jpg, profile-ico.jpg]

// CreateThumbnail generates a thumbnail.
func (gen *Generator) GetProcessedImage(i *Image, dimension ImageDimension) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %v", r)
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// check image validity
	if i == nil || i.ImageData == nil {
		return nil, ErrInvalidImageData
	}

	var mark image.Image
	// check transform valid
	if dimension.Percentage > 0.0 {
		// Resize the image to width = 200px preserving the aspect ratio.
		mark = imgconv.Resize(i.ImageData, &imgconv.ResizeOption{Percent: dimension.Percentage})
	} else if dimension.Width > 0 || dimension.Height > 0 {
		mark = imgconv.Resize(i.ImageData, &imgconv.ResizeOption{Width: dimension.Width, Height: dimension.Height})
	} else {
		return nil, ErrInvalidNoTransformProvided
	}

	return mark, nil
}

// CreateThumbnail generates a thumbnail.
func (gen *Generator) CreateThumbnail(i *Image, dimension ImageDimension) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %v", r)
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// check image validity
	if i == nil || i.ImageData == nil {
		return nil, ErrInvalidImageData
	}

	var mark image.Image
	// check transform valid
	if dimension.Percentage > 0.0 {
		// Resize the image to width = 200px preserving the aspect ratio.
		mark = imgconv.Resize(i.ImageData, &imgconv.ResizeOption{Percent: dimension.Percentage})
	} else if dimension.Width > 0 || dimension.Height > 0 {
		mark = imgconv.Resize(i.ImageData, &imgconv.ResizeOption{Width: dimension.Width, Height: dimension.Height})
	} else {
		return nil, ErrInvalidNoTransformProvided
	}

	return mark, nil
}

// Generate generates all the images for the specified file with the dimensions on the generator
func (gen *Generator) Generate(i *Image) ([]GenerationResult, error) {
	result := make([]GenerationResult, 0)

	//MAYBE: Maybe more specific for this function ?
	if len(gen.OutputFormats) == 0 {
		return nil, ErrInvalidNoTransformProvided
	}

	//
	for _, outputFormat := range gen.OutputFormats {
		thumbImg, err := gen.GetProcessedImage(i, outputFormat)
		if err != nil {
			result = append(result, GenerationResult{
				Filename: i.Path,
				Path:     i.Path,
				Error:    err,
			})
			//return nil, err
			continue
		}

		img := i
		img.ImageData = thumbImg

		save, err := gen.Save(img)
		if err != nil {
			result = append(result, GenerationResult{
				Filename: i.Path,
				Path:     i.Path,
				Error:    err,
			})
			continue
		}

		result = append(result, save)
	}

	return result, nil
}

// CreateThumbnail generates a thumbnail.
func (gen *Generator) Save(i *Image) (result GenerationResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %v", r)
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// check image validity
	if i == nil || i.ImageData != nil {
		return GenerationResult{}, ErrInvalidImageData
	}

	basefileName := filepath.Base(i.Path)
	directoryPath := gen.DestinationPath + gen.DestinationPrefix
	destpath := filepath.Join(directoryPath, basefileName)

	// Write the resulting image as TIFF.
	if err := imgconv.Save(destpath, i.ImageData, &gen.PreferredFormat); err != nil {
		log.Printf("failed to write image: %v", err)
		return GenerationResult{}, fmt.Errorf("failed to write image: %v", err)
	}

	return GenerationResult{
		Filename: basefileName,
		Path:     destpath,
	}, nil
}
