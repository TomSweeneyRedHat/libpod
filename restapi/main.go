package main

// To Run: ./libpod/bin/restapi
// To Test: curl {ip}:8080/images # Get Image numbers
// To Test: curl 192.168.122.243:8080/image?id=599aae32efb4 # Get Image by ID
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/containers/storage/pkg/reexec"
	"github.com/gorilla/mux"
	digest "github.com/opencontainers/go-digest"
	"github.com/projectatomic/libpod/cmd/podman/formats"
	"github.com/projectatomic/libpod/libpod"
	"github.com/projectatomic/libpod/libpod/image"
)

var runtime *libpod.Runtime

//var runtime *image.Runtime

// Possible Util definition

type imagesTemplateParams struct {
	Repository  string
	Tag         string
	ID          string
	Digest      digest.Digest
	Created     string
	CreatedTime time.Time
	Size        string
}

type imagesJSONParams struct {
	ID      string        `json:"id"`
	Name    []string      `json:"names"`
	Digest  digest.Digest `json:"digest"`
	Created time.Time     `json:"created"`
	Size    *uint64       `json:"size"`
}

// End Possible Util definition

func main() {
	if reexec.Init() {
		return
	}

	if runtime == nil {
		getRuntime()
	}
	fmt.Println("LibPod Rest API Hello")
	router := mux.NewRouter()
	router.HandleFunc("/images", GetImages).Methods("GET")
	router.HandleFunc("/image", GetImage).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func GetImage(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("id")
	//input:="599aae32efb4"
	image, err := runtime.ImageRuntime().NewFromLocal(input)
	if err != nil {
		result := fmt.Sprintf("unable to get image, %v", err)
		json.NewEncoder(w).Encode(result)
		return
	}

	result := fmt.Sprintf("TOM: Get Image: %v", image)
	json.NewEncoder(w).Encode(result)
}

func GetImages(w http.ResponseWriter, r *http.Request) {
	//        images, err := runtime.GetImageResults()
	images, err := runtime.ImageRuntime().GetImages()
	if err != nil {
		result := fmt.Sprintf("unable to get images, %v", err)
		json.NewEncoder(w).Encode(result)
		return
	}
	ctx := context.TODO()
	if len(images) == 0 {
		//		return nil
		return
	}
	var out formats.Writer

	imagesOutput := getImagesJSONOutput(ctx, runtime, images)
	out = formats.JSONStructArray{Output: imagesToGeneric([]imagesTemplateParams{}, imagesOutput)}
	finalOut := formats.Writer(out).Out()
	//result := fmt.Sprintf("TOM: GetImages: %v", finalOut)
	result := fmt.Sprintf("TOM: GetImages: %v", formats.Writer(out).Out())
	//result := fmt.Sprintf("TOM: GetImages: %v", images)
	json.NewEncoder(w).Encode(result)
}

// Possible util code:

// getImagesJSONOutput returns the images information in its raw form
func getImagesJSONOutput(ctx context.Context, runtime *libpod.Runtime, images []*image.Image) (imagesOutput []imagesJSONParams) {
	for _, img := range images {
		size, err := img.Size(ctx)
		if err != nil {
			size = nil
		}
		params := imagesJSONParams{
			ID:      img.ID(),
			Name:    img.Names(),
			Digest:  img.Digest(),
			Created: img.Created(),
			Size:    size,
		}
		imagesOutput = append(imagesOutput, params)
	}
	return
}

// imagesToGeneric creates an empty array of interfaces for output
func imagesToGeneric(templParams []imagesTemplateParams, JSONParams []imagesJSONParams) (genericParams []interface{}) {
	if len(templParams) > 0 {
		for _, v := range templParams {
			genericParams = append(genericParams, interface{}(v))
		}
		return
	}
	for _, v := range JSONParams {
		genericParams = append(genericParams, interface{}(v))
	}
	return
}

// End Possible util code

func getRuntime() {
	var err error
	fmt.Printf("TOM blank options\n")
	options := []libpod.RuntimeOption{}
	//	storageOpts := storage.DefaultStoreOptions
	//	options = append(options, libpod.WithStorageConfig(storageOpts))
	runtime, err = libpod.NewRuntime(options...)

	// An empty StoreOptions will use defaults, which is /var/lib/containers/storage
	// If you define GraphRoot and RunRoot to a tempdir, it will create a new storage
	//	storeOpts := storage.StoreOptions{}
	// The first step to using the image library is almost always to create an
	// imageRuntime.
	//	runtime, err = image.NewImageRuntimeFromOptions(storeOpts)

	fmt.Printf("Runtime %v\n", runtime)
	if err != nil {
		result := fmt.Sprintf("unable to get libpod runtime %v\n", err)
		fmt.Printf(result)
		os.Exit(1)
	}
	return
}
