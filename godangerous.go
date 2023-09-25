package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/tidwall/gjson"
)

var cmdr_position string

type location struct {
	station string
	system  string
}

type Planets struct {
	earth_like             []string
	terraform_rocky_body   []string
	terraform_hmetal_world []string
	terraform_water_world  []string
	water_world            []string
	amonia_world           []string
}

func main() {

	var logs string
	var pre_cmdr_position string

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if runtime.GOOS == "windows" {
		logs = home + "\\Saved Games\\Frontier Developments\\Elite Dangerous"
	} else {
		logs = home + "/home/my_username/.local/share/Steam/steamapps/common/Elite Dangerous/Products/elite-dangerous-64/Logs/Saved Games"
	}

	for {
		time.Sleep(1 * time.Second)
		cmdr_position = find_cmdr_position(logs)

		if cmdr_position == "" {
			clear_cli()
			fmt.Println("Waiting FSD jump ...")

		}

		// we dont make http request if commander position dont change
		if cmdr_position != pre_cmdr_position {

			t1 := table.NewWriter()
			t1.SetOutputMirror(os.Stdout)
			t1.AppendHeader(table.Row{"Type of materials", "System", "Station"})
			t1.AppendRows([]table.Row{
				{"Raw", get_trade_raw((cmdr_position)).system, get_trade_raw(cmdr_position).station},
				{"Manufactured", get_trade_manu((cmdr_position)).system, get_trade_manu(cmdr_position).station},
				{"Encoded", get_trade_data((cmdr_position)).system, get_trade_data(cmdr_position).station},
			})
			t1.SetStyle(table.StyleColoredBright)

			t2 := table.NewWriter()
			t2.SetOutputMirror(os.Stdout)
			t2.AppendHeader(table.Row{"Planet type", "Name"})
			t2.AppendRows([]table.Row{
				{"High metal content world terraformable", get_interest_body(cmdr_position).terraform_hmetal_world},
				{"Earth-like world terraformable", get_interest_body(cmdr_position).earth_like},
				{"Rocky Body terraformable", get_interest_body(cmdr_position).terraform_rocky_body},
				{"Water world terraformable", get_interest_body(cmdr_position).terraform_water_world},
				{"Ammonia world", get_interest_body(cmdr_position).amonia_world},
				{"Water world", get_interest_body(cmdr_position).water_world},
			})
			t2.SetStyle(table.StyleColoredBright)

			clear_cli()
			fmt.Println("You are in the system : " + cmdr_position)
			fmt.Println("")
			fmt.Println(" ==== Nearest Traders ====")
			fmt.Println("")
			t1.Render()
			fmt.Println("")
			fmt.Println(" ==== Rare planets in the system ====")
			fmt.Println("")
			t2.Render()

		}
		pre_cmdr_position = cmdr_position

	}

}

func clear_cli() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func find_cmdr_position(folder_path string) string {
	var fsd_jump string
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

		if strings.Contains(f.Name(), "Journal") {
			if f.Name() != "." && f.ModTime().After(newest_time) {
				newest_file = f
				newest_time = f.ModTime()
			}
		}

	}

	file, err := os.Open(folder_path + "\\" + newest_file.Name())
	if err != nil {
		fmt.Println("Unable to open commander log file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "FSDJump") {
			fsd_jump = scanner.Text()

		}
	}

	if fsd_jump != "" {
		cmdr_position := gjson.Get(fsd_jump, "StarSystem")
		return cmdr_position.String()
	} else {
		return ""
	}

}

func get_trade_raw(cmdr_position string) location {
	var trade_raw location
	c := colly.NewCollector()

	visited := false
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		if visited {
			return
		}
		if strings.Contains(e.ChildText("td"), "Pur") {
			trade_raw.station = e.ChildText("strong")
			trade_raw.system = e.ChildText("td > small > a:first-child")
			visited = true

		}

	})

	c.Visit("https://www.edsm.net/fr/search/stations/index/cmdrPosition/" + cmdr_position + "/economy/3/service/71/sortBy/distanceCMDR")

	return trade_raw

}

func get_trade_manu(cmdr_position string) location {
	var trade_manu location
	c := colly.NewCollector()

	visited := false
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		if visited {
			return
		}
		if strings.Contains(e.ChildText("td"), "Manufacturé") {
			trade_manu.station = e.ChildText("strong")
			trade_manu.system = e.ChildText("td > small > a:first-child")
			visited = true

		}

	})

	c.Visit("https://www.edsm.net/fr/search/stations/index/cmdrPosition/" + cmdr_position + "/economy/5/service/71/sortBy/distanceCMDR")

	return trade_manu

}

func get_trade_data(cmdr_position string) location {
	var trade_data location
	c := colly.NewCollector()

	visited := false
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		if visited {
			return
		}
		if strings.Contains(e.ChildText("td"), "Encodé") {
			trade_data.station = e.ChildText("strong")
			trade_data.system = e.ChildText("td > small > a:first-child")
			visited = true

		}

	})

	c.Visit("https://www.edsm.net/fr/search/stations/index/cmdrPosition/" + cmdr_position + "/economy/4/service/71/sortBy/distanceCMDR")

	return trade_data

}

func get_interest_body(cmdr_position string) Planets {
	var planets Planets

	resp, err := http.Get("https://www.edsm.net/api-system-v1/bodies?systemName=" + cmdr_position)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	data := string(body)

	for i := 0; i < 100; i++ {
		subType := gjson.Get(data, "bodies."+strconv.Itoa(i)+".subType")
		terraformingState := gjson.Get(data, "bodies."+strconv.Itoa(i)+".terraformingState")
		name := gjson.Get(data, "bodies."+strconv.Itoa(i)+".name")

		if subType.String() == "High metal content world" && terraformingState.String() == "Candidate for terraforming" {
			planets.terraform_hmetal_world = append(planets.terraform_hmetal_world, name.String())
		} else if subType.String() == "Water world" && terraformingState.String() == "Candidate for terraforming" {
			planets.terraform_water_world = append(planets.terraform_water_world, name.String())
		} else if subType.String() == "Rocky body" && terraformingState.String() == "Candidate for terraforming" {
			planets.terraform_rocky_body = append(planets.terraform_rocky_body, name.String())
		} else if subType.String() == "Earth-like world" {
			planets.earth_like = append(planets.earth_like, name.String())
		} else if subType.String() == "Water world" {
			planets.water_world = append(planets.water_world, name.String())
		} else if subType.String() == "Ammonia world" {
			planets.amonia_world = append(planets.amonia_world, name.String())
		}
	}

	return planets
}
