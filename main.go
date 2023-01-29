package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

func main() {

	println(os.Getenv("google_application_credentials"),
		os.Getenv("target_folder_id"),
		os.Getenv("deploy_item"))

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("google_application_credentials"))

	ctx := context.Background()

	srv, err := drive.NewService(ctx)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	filename := os.Getenv("deploy_item")
	basename := filepath.Base(filename)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	stat, err := file.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	folderid := os.Getenv("target_folder_id")

	res, err := srv.Files.Create(
		&drive.File{
			Parents: []string{folderid},
			Name:    basename,
		},
	).Media(file, googleapi.ChunkSize(int(stat.Size()))).Do()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%s\n", res.Id)

	res2, err := srv.Permissions.Create(res.Id, &drive.Permission{
		Role: "reader",
		Type: "anyone",
	}).Do()

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%s\n", res2.DisplayName)

	r, err := srv.Files.List().SupportsAllDrives(true).PageSize(1000).
		Fields("files(id, name)").
		Q(fmt.Sprintf("'%s' in parents", folderid)). // 特定のフォルダ配下
		Context(ctx).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}

	for _, f := range r.Files {
		println(f.Name, f.Id)
	}

	os.Exit(0)
}
