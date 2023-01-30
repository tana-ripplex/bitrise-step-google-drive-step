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

	filename := os.Getenv("deploy_item")
	folderid := os.Getenv("target_folder_id")

	ctx := context.Background()
	srv, err := drive.NewService(ctx)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	fInfo, _ := os.Stat(filename)

	if !fInfo.IsDir() {
		files, _ := os.ReadDir(filename)

		for _, f := range files {
			path := filepath.Join(filename, f.Name())
			res3, e := upload(path, srv, folderid)
			if !res3 {
				log.Fatalf("Unable to retrieve files: %v", e)
			}
		}
	} else {
		res3, e := upload(filename, srv, folderid)
		if !res3 {
			log.Fatalf("Unable to retrieve files: %v", e)
		}

	}

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

func upload(filename string, srv *drive.Service, folderid string) (bool, error) {

	basename := filepath.Base(filename)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	stat, err := file.Stat()
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	defer file.Close()

	res, err := srv.Files.Create(
		&drive.File{
			Parents: []string{folderid},
			Name:    basename,
		},
	).Media(file, googleapi.ChunkSize(int(stat.Size()))).Do()
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	fmt.Printf("%s\n", res.Id)

	res2, err := srv.Permissions.Create(res.Id, &drive.Permission{
		Role: "reader",
		Type: "anyone",
	}).Do()

	if err != nil {
		log.Fatalln(err)
		return false, err
	}

	fmt.Printf("%s\n", res2.DisplayName)
	return true, nil

}
