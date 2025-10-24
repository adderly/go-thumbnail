// Package thumbnail provides a method to create thumbnails from images.
package thumbnail

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/fs"
	"log"
	"os"
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

	//Name
	DestinationOverride string
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
func New(c Generator) *Generator {
	return &Generator{
		Width:           c.Width,
		Height:          c.Height,
		Name:            c.Name,
		DestinationPath: c.DestinationPath,
		Prefix:          c.Prefix,
		PreferredFormat: imgconv.FormatOption{Format: imgconv.JPEG},
	}
}

// NewGenerator returns an instance of a thumbnail generator with a
// given configuration.
func NewGenerator(c Generator, outputFormats []ImageDimension) *Generator {
	return &Generator{
		Width:           300,
		Height:          300,
		Name:            c.Name,
		DestinationPath: c.DestinationPath,
		Prefix:          c.Prefix,
		PreferredFormat: imgconv.FormatOption{Format: imgconv.JPEG},
		OutputFormats:   outputFormats,
	}
}

// Generator registers a generator configuration to be used when
// creating thumbnails.
type Generator struct {
	// Width is the destination thumbnail width.
	Width int

	// Height is the destination thumbnail height.
	Height int

	// The preferred format for exporting the thumbnails
	PreferredFormat imgconv.FormatOption

	//Name is the game it will output ass
	Name string

	// DestinationPath is the destination thumbnail path.
	DestinationPath string

	// Prefix is the prefix for the destination thumbnail
	// filename.
	Prefix string

	// OutputFormats the formats (dimensions), that the image will be exported to.
	OutputFormats []ImageDimension
}

// GetGeneratorDimension return a dimension object based on the values inside the generator.
func (gen *Generator) GetGeneratorDimension() ImageDimension {
	return ImageDimension{
		Width:  gen.Width,
		Height: gen.Height,
	}
}

// NewImageFromFile reads in an image file from the file system and
// populates an Image object. That new Image object is returned along
// with any errors that occur during the operation.
func (gen *Generator) NewImageFromFile(path string) (*Image, error) {
	// Open a test image.
	img, err := ImageFromFile(path)
	if err != nil {
		return nil, err
	}

	img.TargetDimension = ImageSize{
		Width:  gen.Width,
		Height: gen.Height,
	}

	return img, nil
}

// NewImageFromFile reads in an image file from the file system and
// populates an Image object. That new Image object is returned along
// with any errors that occur during the operation.
func (gen *Generator) NewImageFromFilewWithDefault(path string, defaultImg string) (*Image, error) {
	// Open a test image.
	img, err := ImageFromFile(path)
	if err != nil {
		if defaultImg != "" {
			img, err = ImageFromFile(defaultImg)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	img.TargetDimension = ImageSize{
		Width:  gen.Width,
		Height: gen.Height,
	}

	return img, nil
}

// NewImageFromByteArray reads in an image file from the file system and
// populates an Image object. That new Image object is returned along
// with any errors that occur during the operation.
func (gen *Generator) NewImageFromByteArray(path []byte) (*Image, error) {
	// Open a test image.
	// This should not crash the program
	src, err := imgconv.Decode(bytes.NewBuffer(path))
	if err != nil {
		log.Printf("failed to open image: %v", err)
		return nil, err
	}

	return &Image{
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

// Generate an image, Profile [profile-xl , profile-sm, profile-ico]
// Source: Profile.png
// Generated Result: [ profile-xl.jpg ,  profile-sm.jpg, profile-ico.jpg]

// GetProcessedImage get the processed image from resize.
func (gen *Generator) GetProcessedImage(i *Image, dimension ImageDimension) (img image.Image, err error) {

	return CreateThumbnail(i, dimension)
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

		save, err := gen.SaveWithDimension(img, &outputFormat)
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

// Save save the image
func (gen *Generator) Save(i *Image) (result GenerationResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %v", r)
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// check image validity
	if i == nil || i.ImageData == nil {
		return GenerationResult{}, ErrInvalidImageData
	}

	//get different naming from Image or Generator

	basefileName := filepath.Base(i.Path)
	if len(gen.Name) > 0 {
		basefileName = gen.Name
	}

	directoryPath := gen.DestinationPath
	destpath := filepath.Join(directoryPath, gen.Prefix+basefileName)

	// Write the resulting image as TIFF.
	if err := saveInternal(destpath, i.ImageData, &gen.PreferredFormat); err != nil {
		log.Printf("failed to write image: %v", err)
		return GenerationResult{}, fmt.Errorf("failed to write image: %v", err)
	}

	return GenerationResult{
		Filename: basefileName,
		Path:     destpath,
	}, nil
}

func saveInternal(output string, base image.Image, option *imgconv.FormatOption) error {
	// try to save
	alreadyTried := false
try_again:
	// Write the resulting image as TIFF.
	if err := imgconv.Save(output, base, option); err != nil {
		pathErr := err.(*fs.PathError)
		if pathErr != nil && !alreadyTried {
			alreadyTried = true
			// make dir
			dirPath := filepath.Dir(output)
			mkdirErr := os.MkdirAll(dirPath, 0755)
			if mkdirErr != nil {
			}
			goto try_again
		}

		log.Printf("failed to write image: %v", err)
		return err
	}
	return nil
}

// SaveWithDimension generates a thumbnail.
func (gen *Generator) SaveWithDimension(i *Image, imgConf *ImageDimension) (result GenerationResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %v", r)
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// check image validity
	if i == nil || i.ImageData == nil || imgConf == nil {
		return GenerationResult{}, ErrInvalidImageData
	}

	//get different naming from Image or Generator

	var prefix string
	var basefileName string
	var directoryPath string
	var destpath string

	if len(imgConf.Prefix) > 0 {
		prefix = imgConf.Prefix
	} else {
		prefix = gen.Prefix
	}

	if len(imgConf.Name) > 0 {
		basefileName = filepath.Base(imgConf.Name)
	} else {
		basefileName = filepath.Base(i.Path)
	}

	if len(imgConf.DestinationOverride) > 0 {
		destpath = imgConf.DestinationOverride
	} else {
		directoryPath = gen.DestinationPath
	}

	fileLocationPath := filepath.Join(directoryPath, prefix+basefileName)

	//try_again:
	// Write the resulting image as TIFF.
	if err := saveInternal(fileLocationPath, i.ImageData, &gen.PreferredFormat); err != nil {
		log.Printf("failed to write image: %v", err)
		return GenerationResult{}, fmt.Errorf("failed to write image: %v", err)
	}

	return GenerationResult{
		Filename: basefileName,
		Path:     destpath,
	}, nil
}

//http://localhost:9999/resource/gen?Id=27

func ImageFromFile(path string) (*Image, error) {
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
		TargetDimension: DefaultThumbnailSize,
	}, nil
}

// CreateThumbnail generates a thumbnail.
func CreateThumbnail(i *Image, dimension ImageDimension) (img image.Image, err error) {
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
	} else if dimension.Width > 0 && dimension.Height > 0 {
		mark = imgconv.Resize(i.ImageData, &imgconv.ResizeOption{Width: dimension.Width, Height: dimension.Height})
	} else if dimension.Width > 0 {
		mark = imgconv.Resize(i.ImageData, &imgconv.ResizeOption{Width: dimension.Width})
	} else if dimension.Height > 0 {
		mark = imgconv.Resize(i.ImageData, &imgconv.ResizeOption{Height: dimension.Height})
	} else {
		return nil, ErrInvalidNoTransformProvided
	}

	return mark, nil
}

// SaveRaw generates a thumbnail.
func SaveRaw(i image.Image, path string, format imgconv.FormatOption) (result GenerationResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered from panic: %v", r)
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	// check image validity
	if i == nil {
		return GenerationResult{}, ErrInvalidImageData
	}
	if len(path) == 0 {
		return GenerationResult{}, ErrInvalidImageData
	}

	//get different naming from Image or Generator

	basefileName := filepath.Base(path)
	destpath := path

	//try_again:
	// Write the resulting image as TIFF.
	if err := imgconv.Save(destpath, i, &format); err != nil {
		log.Printf("failed to write image: %v", err)
		return GenerationResult{}, fmt.Errorf("failed to write image: %v", err)
	}

	return GenerationResult{
		Filename: basefileName,
		Path:     destpath,
	}, nil
}
