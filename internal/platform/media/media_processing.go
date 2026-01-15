package media

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	DefaultSize = 128
	MinSize     = 16
	MaxSize     = 1024
)

var (
	// ErrCWebPUnavailable indicates cwebp is missing from PATH.
	ErrCWebPUnavailable = errors.New("cwebp not available for WebP encoding")
	// ErrImageTooSmall indicates the source image is smaller than the target size.
	ErrImageTooSmall = errors.New("source image smaller than requested size")
)

// WebPOptions customizes cwebp behavior.
type WebPOptions struct {
	Width     int
	Height    int
	Quality   int
	SizeLimit int
	Method    int
	ExtraArgs []string
}

// EncodeWebP converts or resizes an image using cwebp.
func EncodeWebP(inputPath, outputPath string, width, height, quality int) error {
	return EncodeWebPWithOptions(inputPath, outputPath, WebPOptions{
		Width:   width,
		Height:  height,
		Quality: quality,
	})
}

// EncodeWebPWithOptions converts an image using cwebp with custom options.
func EncodeWebPWithOptions(inputPath, outputPath string, opts WebPOptions) error {
	cwebpPath, err := exec.LookPath("cwebp")
	if err != nil {
		return ErrCWebPUnavailable
	}
	args := []string{"-quiet"}
	if opts.Width > 0 && opts.Height > 0 {
		args = append(args, "-resize", strconv.Itoa(opts.Width), strconv.Itoa(opts.Height))
	}
	if opts.Quality > 0 {
		args = append(args, "-q", strconv.Itoa(opts.Quality))
	}
	if opts.SizeLimit > 0 {
		args = append(args, "-size", strconv.Itoa(opts.SizeLimit))
	}
	if opts.Method > 0 {
		args = append(args, "-m", strconv.Itoa(opts.Method))
	}
	if len(opts.ExtraArgs) > 0 {
		args = append(args, opts.ExtraArgs...)
	}
	args = append(args, inputPath, "-o", outputPath)
	cmd := exec.Command(cwebpPath, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		details := strings.TrimSpace(string(output))
		if details != "" {
			return fmt.Errorf("cwebp failed: %w: %s", err, details)
		}
		return fmt.Errorf("cwebp failed: %w", err)
	}
	return nil
}

// EnsureDefaultWebP creates a default webp profile picture from the png if missing.
func EnsureDefaultWebP(staticDir string) error {
	pngPath := filepath.Join(staticDir, "img", "default_profile_picture.png")
	webpPath := filepath.Join(staticDir, "img", "default_profile_picture.webp")
	if _, err := os.Stat(webpPath); err == nil {
		return nil
	}
	if _, err := os.Stat(pngPath); err != nil {
		return nil
	}
	if err := EncodeWebP(pngPath, webpPath, DefaultSize, DefaultSize, 80); err != nil {
		return err
	}
	return nil
}

// WriteWebP decodes the uploaded image and writes a WebP version to disk.
func WriteWebP(src io.Reader, destPath string) error {
	tmpInput, err := os.CreateTemp(filepath.Dir(destPath), "profile_picture_*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpInput.Name())
	}()
	if _, err := io.Copy(tmpInput, src); err != nil {
		_ = tmpInput.Close()
		return err
	}
	if err := tmpInput.Close(); err != nil {
		return err
	}

	tmpOutput := destPath + ".tmp"
	targetW, targetH, err := profilePictureResize(tmpInput.Name(), 1024)
	if err != nil {
		_ = os.Remove(tmpOutput)
		return err
	}
	if err := EncodeWebPWithOptions(tmpInput.Name(), tmpOutput, WebPOptions{
		Width:     targetW,
		Height:    targetH,
		SizeLimit: 400000,
		Method:    6,
	}); err != nil {
		_ = os.Remove(tmpOutput)
		return err
	}
	return os.Rename(tmpOutput, destPath)
}

func profilePictureResize(path string, maxDim int) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	w, h := cfg.Width, cfg.Height
	if w == 0 || h == 0 {
		return 0, 0, errors.New("invalid image dimensions")
	}
	if w <= maxDim && h <= maxDim {
		return w, h, nil
	}
	if w >= h {
		newW := maxDim
		newH := int(float64(h) * float64(maxDim) / float64(w))
		if newH < 1 {
			newH = 1
		}
		return newW, newH, nil
	}
	newH := maxDim
	newW := int(float64(w) * float64(maxDim) / float64(h))
	if newW < 1 {
		newW = 1
	}
	return newW, newH, nil
}

// ResizeAndCache scales an image to size and caches it as WebP.
func ResizeAndCache(sourcePath, cachePath string, size int) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		return err
	}
	srcW := srcImg.Bounds().Dx()
	srcH := srcImg.Bounds().Dy()
	if srcW < size && srcH < size {
		return ErrImageTooSmall
	}
	targetW, targetH := FitCover(srcW, srcH, size)

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), srcImg, srcImg.Bounds(), draw.Over, nil)

	tmpPNG, err := os.CreateTemp(filepath.Dir(cachePath), "profile_picture_cache_*.png")
	if err != nil {
		return err
	}
	tmpPNGPath := tmpPNG.Name()
	if err := png.Encode(tmpPNG, dst); err != nil {
		_ = tmpPNG.Close()
		_ = os.Remove(tmpPNGPath)
		return err
	}
	if err := tmpPNG.Close(); err != nil {
		_ = os.Remove(tmpPNGPath)
		return err
	}
	defer func() {
		_ = os.Remove(tmpPNGPath)
	}()

	tmpOutput := cachePath + ".tmp"
	if err := EncodeWebP(tmpPNGPath, tmpOutput, targetW, targetH, 80); err != nil {
		_ = os.Remove(tmpOutput)
		return err
	}
	return os.Rename(tmpOutput, cachePath)
}

func FitWithin(width, height, maxDim int) (int, int) {
	if width <= 0 || height <= 0 || maxDim <= 0 {
		return maxDim, maxDim
	}
	if width <= maxDim && height <= maxDim {
		return width, height
	}
	if width >= height {
		newW := maxDim
		newH := int(float64(height) * float64(maxDim) / float64(width))
		if newH < 1 {
			newH = 1
		}
		return newW, newH
	}
	newH := maxDim
	newW := int(float64(width) * float64(maxDim) / float64(height))
	if newW < 1 {
		newW = 1
	}
	return newW, newH
}

func FitCover(width, height, minDim int) (int, int) {
	if width <= 0 || height <= 0 || minDim <= 0 {
		return minDim, minDim
	}
	if width >= minDim && height >= minDim {
		if width <= height {
			newW := minDim
			newH := int(float64(height) * float64(minDim) / float64(width))
			if newH < 1 {
				newH = 1
			}
			return newW, newH
		}
		newH := minDim
		newW := int(float64(width) * float64(minDim) / float64(height))
		if newW < 1 {
			newW = 1
		}
		return newW, newH
	}
	if width < height {
		newW := minDim
		newH := int(float64(height) * float64(minDim) / float64(width))
		if newH < minDim {
			newH = minDim
		}
		return newW, newH
	}
	newH := minDim
	newW := int(float64(width) * float64(minDim) / float64(height))
	if newW < minDim {
		newW = minDim
	}
	return newW, newH
}
