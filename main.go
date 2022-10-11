package main

import (
	"io/ioutil"
	"log"
	"mt-logs/readlogs"
)

func main() {
	mymap := make(map[string][]readlogs.JobBackup)

	list := readlogs.GetlistJobs("./Svc.VeeamBackup.log")
	//for _, job := range list {

	// for each file in the directory
	files, err := ioutil.ReadDir(list[0])
	if err != nil {
		log.Fatal(err)
	}
	a := []readlogs.JobBackup{}
	for _, f := range files {
		if f.IsDir() {
			j := readlogs.JobBackup{Server: f.Name()}
			a = append(a, j)
		}
	}
	mymap[list[0]] = a
	for _, mymap := range mymap {
		for _, j := range mymap {
			readlogs.GetJobBackup("./Svc.VeeamBackup.log", j)
		}
	}

}
