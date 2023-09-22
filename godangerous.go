package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"runtime"
	"time"
)

type location struct {
	station string
	system string
}

type planets struct {
	eath_like []string{}
	terraform_rocky_body []string{}
	terraform_hmetal_world []string{}
	terraform_water_world []string{}
	water_world []string{}
	amonia_world []string{}
}


func main() {

	var logs string

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if runtime.GOOS == "windows" {
		logs = home + "\\Saved Games\\Frontier Developments\\Elite Dangerous"
	} else {
		logs = "/home/my_username/.local/share/Steam/steamapps/common/Elite Dangerous/Products/elite-dangerous-64/Logs/Saved Games"
	}

	fmt.Println(logs)

	find_newest_file(logs)
}

func find_newest_file(folder_path string) string {
	var list_of_file []string
	fs.WalkDir(os.DirFS(folder_path), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		list_of_file = append(list_of_file, folder_path+"\\"+path)
		return nil
	})

	var newest_file os.FileInfo
	var newest_time time.Time

	for _, file := range list_of_file {
		f, err := os.Stat(file)
		if err != nil {
			log.Fatal(err)
		}

		if f.Name() != "." && f.ModTime().After(newest_time) {
			newest_file = f
			newest_time = f.ModTime()
		}

	}

	return newest_file.Name()

}
