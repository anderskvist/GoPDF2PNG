package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"time"

	"gopkg.in/gographics/imagick.v2/imagick"

	"github.com/anderskvist/GoHelpers/log"
	"github.com/anderskvist/GoHelpers/version"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	log.Info("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("PDF")
	if err != nil {
		log.Error("Error Retrieving the File")
		log.Error(err)
		return
	}
	defer file.Close()

	log.Debugf("Uploaded File: %+v\n", handler.Filename)
	log.Debugf("File Size: %+v\n", handler.Size)
	log.Debugf("MIME Header: %+v\n", handler.Header)

	tempFile, err := ioutil.TempFile("/tmp/", "upload-*")
	if err != nil {
		log.Error(err)
		return
	}
	defer tempFile.Close()

	outputFile, err := ioutil.TempFile("/tmp/", "output-*.png")
	if err != nil {
		log.Error(err)
		return
	}
	defer outputFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(err)
		return
	}

	// write this byte array to our temporary file
	_, err = tempFile.Write(fileBytes)
	if err != nil {
		log.Error(err)
		return
	} else {
		if err := ConvertPdfToJpg(tempFile.Name(), outputFile.Name()); err != nil {
			log.Fatal(err)
			return
		}

		downloadBytes, err := ioutil.ReadFile(outputFile.Name())

		if err != nil {
			log.Error(err)
			return
		}

		mime := http.DetectContentType(downloadBytes)
		fileSize := len(string(downloadBytes))

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Content-Disposition", "attachment; filename="+path.Base(outputFile.Name())+"")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Content-Length", strconv.Itoa(fileSize))
		w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

		http.ServeContent(w, r, outputFile.Name(), time.Now(), bytes.NewReader(downloadBytes))

		//Everything went well, so we redirect to frontpage
		http.Redirect(w, r, "/", 301)
	}
}

func main() {
	log.Infof("GoPDF2PNG version: %s.\n", version.Version)

	fs := http.FileServer(http.Dir("html"))
	http.Handle("/", fs)

	http.HandleFunc("/upload", uploadFile)
	log.Fatal(http.ListenAndServe(":80", nil))
}

// ConvertPdfToJpg will take a filename of a pdf file and convert the file into an
// image which will be saved back to the same location. It will save the image as a
// high resolution jpg file with minimal compression.
func ConvertPdfToJpg(pdfName string, imageName string) error {

	// Setup
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// Must be *before* ReadImageFile
	// Make sure our image is high quality
	if err := mw.SetResolution(300, 300); err != nil {
		return err
	}

	// Load the image file into imagick
	if err := mw.ReadImage(pdfName); err != nil {
		return err
	}

	// Must be *after* ReadImageFile
	// Flatten image and remove alpha channel, to prevent alpha turning black in jpg
	if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_FLATTEN); err != nil {
		return err
	}

	// Set any compression (100 = max quality)
	if err := mw.SetCompressionQuality(95); err != nil {
		return err
	}

	// Select only first page of pdf
	mw.SetIteratorIndex(0)

	// Convert into JPG
	if err := mw.SetFormat("jpg"); err != nil {
		return err
	}

	// Save File
	return mw.WriteImage(imageName)
}
