package readlogs

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"time"
)

type JobBackup struct {
	VeeamServer   string        `json:"veeam_server"`
	Server        string        `json:"server"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	BacupDuration time.Duration `json:"backup_duration"`
	JobID         string        `json:"job_id"`
	JobName       string        `json:"job_name"`
	JobMode       string        `json:"job_mode"`
	JobSessionID  string        `json:"job_session_id"`
	JobCompleted  bool          `json:"job_completed"`
	Status        string        `json:"status"`
	JobSize       string        `json:"job_size"`
}

func filtro(localizar, apagar, texto string) string {
	a, _ := regexp.Compile(localizar)
	b, _ := regexp.Compile(apagar)
	c := b.ReplaceAllString((a.FindString(texto)), "$1")
	return c
}

func GetLogJobs(filepath string) []JobBackup {

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	jobs := []JobBackup{}
	timeparse := "02.01.2006 15:04:05"

	jobstart, _ := regexp.Compile(`^Starting new log`)
	veeamServer, _ := regexp.Compile(`^MachineName: `)
	startTime, _ := regexp.Compile(`Job event 'started' was disposed.`)
	jobID, _ := regexp.Compile(`Job ID:`)
	jobName, _ := regexp.Compile(`Job Name:`)
	jobStatus, _ := regexp.Compile(`Job session '[a-z0-9]+-[a-z0-9]+-[a-z0-9]+-[a-z0-9]+-[a-z0-9]+'`)

	newjob := bool(false)
	j := JobBackup{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if jobstart.MatchString(line) {
			newjob = true
			j = JobBackup{}

		}

		if newjob {

			if veeamServer.MatchString(line) {
				j.VeeamServer = filtro(`MachineName: \[[a-zA-Z0-9._-]+`, `MachineName: \[`, line)
			}

			if startTime.MatchString(line) {
				j.StartTime, _ = time.Parse(timeparse, filtro(`\[[0-9.]+ [0-9:]+`, `\[`, line))
				j.JobSessionID = filtro(`[a-z0-9]+-[a-z0-9]+-[a-z0-9]+-[a-z0-9]+-[a-z0-9]+`, ``, line)
			}

			if jobID.MatchString(line) {
				j.JobID = filtro(`Job ID: \[[a-z0-9-]+`, `Job ID: \[`, line)
			}

			if jobName.MatchString(line) {
				j.JobName = filtro(`Job Name: \[[a-zA-Z0-9._ -]+`, `Job Name: \[`, line)
			}

			if jobStatus.MatchString(line) {
				j.EndTime, _ = time.Parse(timeparse, filtro(`\[[0-9. ]+[0-9:]+`, `\[`, line))
				j.BacupDuration = j.EndTime.Sub(j.StartTime)
				j.Status = filtro(`status: \'[a-zA-Z]+`, `status: \'`, line)
				j.JobSize = filtro(`of \'[0-9A-Z ]+\' bytes`, `\' bytes`, line)
				j.JobSize = filtro(`of \'[0-9A-Z ]+`, `of \'`, j.JobSize)
				j.JobCompleted = true
				j.Server = "192.168.5.40"
				jobs = append(jobs, j)
				newjob = false
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	//body, _ := json.Marshal(jobs)
	//fmt.Println(string(body))
	return jobs
}

// Lista de JobsName
func GetlistJobs(path string) []string {
	var joblist []string

	dateYesterday := time.Now().AddDate(0, 0, -1)
	type Job struct {
		JobName string    `json:"job_name"`
		Date    time.Time `json:"date"`
	}

	list := make(map[string]Job)

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	timeparse := "02.01.2006 15:04:05"

	jobstart, _ := regexp.Compile(`==  Name: \[`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if jobstart.MatchString(line) {
			jobname := filtro(`Name: \[[a-zA-Z0-9. _-]+`, `Name: \[`, line)
			date, _ := time.Parse(timeparse, filtro(`\[[0-9.: ]+`, `\[`, line))
			list[jobname] = Job{JobName: jobname, Date: date}
		}
	}
	for _, l := range list {
		if !l.Date.Before(dateYesterday) {
			joblist = append(joblist, l.JobName)
		}
	}
	regexp, _ := regexp.Compile(` `)
	for k, j := range joblist {
		if regexp.MatchString(j) {
			joblist[k] = regexp.ReplaceAllString(j, "_")
		}
	}
	return joblist
}
