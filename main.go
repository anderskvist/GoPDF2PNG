package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/gographics/imagick.v2/imagick"

	"github.com/anderskvist/GoHelpers/log"
	"github.com/anderskvist/GoHelpers/version"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	log.Info("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
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

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("/tmp/", "upload-*")
	if err != nil {
		log.Error(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Error(err)
	}

	// write this byte array to our temporary file
	_, err = tempFile.Write(fileBytes)
	if err != nil {
		log.Error(err)
	} else {
		if err := ConvertPdfToJpg(tempFile.Name(), "/tmp/testing.png"); err != nil {
			log.Fatal(err)
		}

		downloadBytes, err := ioutil.ReadFile("/tmp/testing.png")

		if err != nil {
			fmt.Println(err)
		}

		// set the default MIME type to send
		mime := http.DetectContentType(downloadBytes)

		fileSize := len(string(downloadBytes))

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Content-Disposition", "attachment; filename=pdf.png")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Content-Length", strconv.Itoa(fileSize))
		w.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")

		// force it down the client's.....
		http.ServeContent(w, r, "/tmp/testing.png", time.Now(), bytes.NewReader(downloadBytes))

	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
	http.Redirect(w, r, "/", 301)

}

func main() {
	log.Infof("GoPDF2PNG version: %s.\n", version.Version)

	pdfName := "test.pdf"
	imageName := "test.jpg"

	fs := http.FileServer(http.Dir("html"))
	http.Handle("/", fs)

	http.HandleFunc("/upload", uploadFile)
	log.Fatal(http.ListenAndServe(":80", nil))

	if err := ConvertPdfToJpg(pdfName, imageName); err != nil {
		log.Fatal(err)
	}
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
