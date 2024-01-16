package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
)

func myLogger(messageDomain string, verbosity vips.LogLevel, message string) {
	var messageLevelDescription string
	switch verbosity {
	case vips.LogLevelError:
		messageLevelDescription = "error"
	case vips.LogLevelCritical:
		messageLevelDescription = "critical"
	case vips.LogLevelWarning:
		messageLevelDescription = "warning"
	case vips.LogLevelMessage:
		messageLevelDescription = "message"
	case vips.LogLevelInfo:
		messageLevelDescription = "info"
	case vips.LogLevelDebug:
		messageLevelDescription = "debug"
	}

	log.Printf("[%v.%v] %v", messageDomain, messageLevelDescription, message)
}
func main() {

	vips.LoggingSettings(myLogger, vips.LogLevelError)
	defer vips.Shutdown()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "" || r.URL.Path == "/" || r.URL.Path == "/favicon.ico" {
			http.Error(w, "invaild url", http.StatusBadRequest)
			return
		}
		width, _ := strconv.Atoi(r.URL.Query().Get("w"))
		height, _ := strconv.Atoi(r.URL.Query().Get("h"))
		quality, _ := strconv.Atoi(r.URL.Query().Get("q"))

		if quality == 0 {
			quality = 75
		}

		params := url.Values{}
		for k, v := range r.URL.Query() {
			if k != "w" && k != "h" && k != "q" {
				params.Add(k, v[0])
			}
		}

		imagePath := "https:/" + r.URL.Path

		query := params.Encode()

		if query != "" {
			imagePath = imagePath + "?" + query
		}

		resp, err := http.Get(imagePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		image, err := vips.NewImageFromReader(resp.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = image.AutoRotate()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		aspectRatio := float32(image.Width()) / float32(image.Height())

		if width == 0 && height != 0 {
			width = int(float32(height) * aspectRatio)
		} else if height == 0 && width != 0 {
			height = int(float32(width) / aspectRatio)
		}

		if width != 0 && height != 0 {
			err := image.Thumbnail(width, height, vips.InterestingCentre)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		imagebytes, _, err := image.ExportWebp(&vips.WebpExportParams{
			Quality:         quality,
			Lossless:        false,
			NearLossless:    false,
			ReductionEffort: 4,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/webp")
		w.Header().Set("Content-Length", strconv.Itoa(len(imagebytes)))
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Write(imagebytes)
	})

	port := "7860"

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	http.ListenAndServe(":"+port, nil)
}
