package sd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// WalkDir walks the directory tree rooted at root, calling walkFn for each
func WalkDir(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if info.IsDir() {
			// Send say hello to endpoint server
			if err := sendSayHello(); err != nil {
				log.Println("Say hello to endpoint server failed\nERROR:", err)
				return err
			}
			f, err := os.Open(root)
			if err != nil {
				log.Printf("open dir %s failed\treason:%v", root, err)
				return err
			}
			files, err := f.Readdir(-1)
			f.Close()
			if err != nil {
				log.Fatal(err)
			}
			for _, file := range files {
				fmt.Println(file.Name())
			}
		}
		return nil
	})
}

// Send a request to say hello to endpoint server
func sendSayHello() error {
	getreq, err := http.Get("http://localhost:8080/sayhello")
	if err != nil {
		log.Println(err)
		return err
	}
	defer getreq.Body.Close()
	log.Printf("Response: %s", getreq.Status)
	return nil
}