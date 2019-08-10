package imagefetcher

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/bmp"
)

// Fetcher struct
type Fetcher struct {
	urls   []string
	images []Image
	sem    int
}

// Image image
type Image struct {
	data      []byte
	imagetype string
	url       string
	filename  string
}

// New Fetcher
func New(url string) (*Fetcher, error) {
	urlSclice := make([]string, 1)
	urlSclice[0] = url
	body := make([]Image, 1)
	fetcher := &Fetcher{
		urls:   urlSclice,
		images: body,
	}
	err := fetcher.fetch(0)
	if err != nil {
		return nil, err
	}
	return fetcher, nil
}

// NewSlice accepts slice of urls
func NewSlice(urls []string, sem int) (*Fetcher, error) {
	body := make([]Image, len(urls))
	fetcher := &Fetcher{
		urls:   urls,
		images: body,
		sem:    sem,
	}

	err := fetcher.fetchAll(sem)
	if err != nil {
		return nil, err
	}

	return fetcher, nil
}

// Fetch Fetch data
func (f *Fetcher) fetchAll(threadCount int) error {
	fmt.Println("Fetching all...")
	sem := make(chan struct{}, int(math.Min(float64(threadCount), float64(len(f.urls)))))

	wg := &sync.WaitGroup{}
	wg.Add(len(f.urls))
	done := func() {
		wg.Done()
		<-sem
	}
	for i := range f.urls {
		sem <- struct{}{}
		// fmt.Println("FATCH")
		go func(index int) {
			err := f.fetch(index)
			filename := filepath.Base(f.urls[index])
			if err != nil {
				fmt.Println("Fetch error: ", filename)
			} else {
				fmt.Println("Fetching DONE: ", filename)
			}
			defer done()
		}(i)
	}
	wg.Wait()
	return nil
}

// Fetch Fetch data
func (f *Fetcher) fetch(urlIndex int) error {
	url := f.urls[urlIndex]
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err
	}

	f.images[urlIndex].data = body
	f.images[urlIndex].imagetype = http.DetectContentType(body)
	f.images[urlIndex].url = f.urls[urlIndex]

	filename := filepath.Base(f.urls[urlIndex])
	baseFilename := strings.TrimSuffix(filename, filepath.Ext(filename))

	f.images[urlIndex].filename = baseFilename

	return nil
}

// SavePng saves fetched image to file
func (f *Fetcher) SavePng(path string) error {
	return nil
}

// SaveAllImagesToDisk saves all images to Disk
func (f *Fetcher) SaveAllImagesToDisk(path string) error {
	for _, image := range f.images {
		filename := filepath.Base(image.url)
		err := image.SaveImageToFile(path)
		if err != nil {
			fmt.Println("Saving error: ", filename)
		} else {
			fmt.Println("Saving DONE: ", filename)
		}
	}
	return nil
}

// SaveAllImagesToDiskInFormat saves all images to Disk in specified format
func (f *Fetcher) SaveAllImagesToDiskInFormat(path string, format string) error {
	for _, image := range f.images {
		err := image.SaveImageToFileInFormat(path, format)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

// SaveImageToFile saves fetched image to file
func (I *Image) SaveImageToFile(saveDir string) error {
	// filename := filepath.Base(I.url)

	switch I.imagetype {
	case "image/png":
		return I.saveImgToPNG(saveDir)
	case "image/jpeg":
		return I.saveImgToJPEG(saveDir)
	case "image/gif":
		return I.saveImgToGIF(saveDir)
	case "image/bmp":
		return I.saveImgToBitmap(saveDir)
	default:
		return fmt.Errorf("Image format not supported")
	}

	return nil
}

// SaveImageToFileInFormat saves fetched image to file in specified format
func (I *Image) SaveImageToFileInFormat(saveDir string, format string) error {
	// filename := filepath.Base(I.url)

	switch format {
	case "png":
		return I.saveImgToPNG(saveDir)
	case "jpeg":
		return I.saveImgToJPEG(saveDir)
	case "gif":
		return I.saveImgToGIF(saveDir)
	case "bmp":
		return I.saveImgToBitmap(saveDir)
	default:
		return fmt.Errorf("Image format not supported")
	}

	return nil
}

// GetAllImages return all images
func (f *Fetcher) GetAllImages() []Image {
	return f.images
}

// GetImage return first image
func (f *Fetcher) GetImage() (Image, error) {
	return f.GetImagebyIndex(0)
}

// GetImageBytes return first image
func (f *Fetcher) GetImageBytes() ([]byte, error) {
	img, err := f.GetImagebyIndex(0)
	return img.data, err
}

// GetImagebyIndex return image by index
func (f *Fetcher) GetImagebyIndex(urlIndex int) (Image, error) {
	if urlIndex >= len(f.images) {
		return Image{}, fmt.Errorf("Index out of range")
	}
	img := f.images[urlIndex]
	return img, nil
}

func (I *Image) saveImgToPNG(path string) error {
	fullOutputPath := []string{path, "/", I.filename, ".png"}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0777)
	}

	img, err := decodeImage(I.data, I.imagetype)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(strings.Join(fullOutputPath, ""), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {

		return err
	}

	return png.Encode(f, img)
}

func (I *Image) saveImgToJPEG(path string) error {
	fullOutputPath := []string{path, "/", I.filename, ".jpeg"}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0777)
	}

	img, err := decodeImage(I.data, I.imagetype)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(strings.Join(fullOutputPath, ""), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	return jpeg.Encode(f, img, nil)
}

func (I *Image) saveImgToGIF(path string) error {
	fullOutputPath := []string{path, "/", I.filename, ".gif"}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0777)
	}

	img, err := decodeImage(I.data, I.imagetype)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(strings.Join(fullOutputPath, ""), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	return gif.Encode(f, img, nil)
}

func (I *Image) saveImgToBitmap(path string) error {
	fullOutputPath := []string{path, "/", I.filename, ".bmp"}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0777)
	}

	img, err := decodeImage(I.data, I.imagetype)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(strings.Join(fullOutputPath, ""), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	return bmp.Encode(f, img)
}

func decodeImage(data []byte, format string) (image.Image, error) {
	format = strings.ToLower(format)
	switch format {
	case "image/png", "png":
		return decodePNG(data)

	case "image/jpeg", "jpeg", "jpg":
		return decodeJPEG(data)

	case "image/gif", "gif":
		return decodeGIF(data)

	case "image/bmp", "bmp", "bitmap":
		return decodeBMP(data)
	default:
		return nil, fmt.Errorf("Format not supported")
	}
}

func decodePNG(data []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(data))
}

func decodeJPEG(data []byte) (image.Image, error) {
	return jpeg.Decode(bytes.NewReader(data))
}

func decodeGIF(data []byte) (image.Image, error) {
	return gif.Decode(bytes.NewReader(data))
}

func decodeBMP(data []byte) (image.Image, error) {
	return bmp.Decode(bytes.NewReader(data))
}
