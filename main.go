package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/felixge/httpsnoop"
	"github.com/rs/cors"
)

func isIPorLocalhost(host string) bool {
	hostname := strings.Split(host, ":")[0]

	if hostname == "localhost" {
		return true
	}
	ip := net.ParseIP(hostname)
	if ip != nil {
		return true
	}

	return false
}
func main() {

	vips.LoggingSettings(nil, vips.LogLevelError)
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

		protocol := "https:"

		if isIPorLocalhost(strings.Split(r.URL.Path, "/")[1]) {
			protocol = "http:"
		}

		imagePath := protocol + "/" + r.URL.Path

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

	handler := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool { return true },
		AllowedMethods:  []string{"GET", "HEAD", "OPTIONS", "POST"},
		MaxAge:          86400,
	}).Handler(http.DefaultServeMux)

	logHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(handler, w, r)
		log.Printf(
			"%s %s (code=%d dt=%s)",
			r.Method,
			r.URL,
			m.Code,
			m.Duration,
		)
	})

	port := "7860"

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: logHandler,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
}
