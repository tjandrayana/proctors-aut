package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tjandrayana/toolings/proctors-aut/workerpool"
)

type Instances struct {
	ServiceName string `json:"service_name"`
	IP          string `json:"ip"`
	Environment string `json:"environment"`
	MachineType string `json:"machine_type"`
	Disk        string `json:"disk"`
	Team        string `json:"team"`
}

func main() {
	RunCommands()
}

func RunCommands() {

	var instances []Instances
	records := readCsvFile("files/input.csv")

	start := time.Now()

	// Start Worker Pool.
	totalWorker := 10
	wp := workerpool.NewWorkerPool(totalWorker)
	wp.Run()

	totalTask := len(records)
	resultC := make(chan Instances, totalTask)

	for i, d := range records {

		wp.AddTask(func() {

			insName := d[0]
			env := d[1]
			team := d[2]

			fmt.Printf("---- Processing Task : %d--- %s -> %s ----- \n", i, insName, env)

			if i > 0 {
				msg := ExecuteCommand(insName, env)
				resultC <- EvaluateOutput(msg, insName, env, team)
			} else {
				resultC <- Instances{}
			}

		})

	}

	for i := 0; i < totalTask; i++ {
		res := <-resultC
		if i == 0 {
			continue
		}
		instances = append(instances, res)
	}

	CreateCSV(instances)

	end := time.Now()

	fmt.Printf("----- Execution -----\nstart on: %v\nend on: %v\nWorker : %d\nduration: %v\n------------------------\n", start, end, totalWorker, time.Since(start))

}

func ExecuteCommand(instanceName, environment string) string {
	args := []string{
		"execute",
		"describe-instance",
		fmt.Sprintf("INSTANCE_NAME=%s", instanceName),
		fmt.Sprintf("ENVIRONMENT=%s", environment),
	}

	msg, err := exec.Command("proctor", args...).Output()
	if err != nil {
		log.Println(err)
	}

	return string(msg)
}

func EvaluateOutput(msg, insName, environment, team string) Instances {

	tables := make(map[string]string)

	m := strings.ToLower(msg)
	resps := strings.Split(m, "\n")

	for _, d := range resps {
		if d == "" {
			continue
		}

		sp := strings.Split(d, ":")
		if len(sp) < 2 {
			continue
		}
		tables[sp[0]] = sp[1]
	}

	for k, v := range tables {
		fmt.Printf("k:%s -> v:%s\n", k, v)
	}

	disk := tables["disk1 size"]

	if tables["disk2 size"] != "" {
		disk = tables["disk2 size"]
	}

	return Instances{
		ServiceName: insName,
		IP:          tables["ip"],
		Environment: environment,
		MachineType: tables["machine type"],
		Disk:        disk,
		Team:        team,
	}

}

func CreateCSV(ins []Instances) {

	var rows [][]string

	file, err := os.Create(fmt.Sprintf("files/detail_instances-%v.csv", time.Now()))
	if err != nil {
		log.Println("Cannot create CSV file:", err)
	}
	defer file.Close()

	rows = append(rows, []string{
		"ServiceName",
		"IP",
		"Environment",
		"MachineType",
		"Disk",
		"Team",
	})

	for _, record := range ins {
		row := []string{
			record.ServiceName,
			record.IP,
			record.Environment,
			record.MachineType,
			record.Disk,
			record.Team,
		}
		rows = append(rows, row)
	}

	writer := csv.NewWriter(file)

	err = writer.WriteAll(rows)
	if err != nil {
		log.Println("Cannot write to CSV file:", err)
	}
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}
