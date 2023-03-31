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
	DiskName1   string `json:"disk_name_1"`
	Disk1Size1  string `json:"disk_size_1"`
	DiskName2   string `json:"disk_name_2"`
	Disk1Size2  string `json:"disk_size_2"`
	Team        string `json:"team"`
}

type DeleteVM struct {
	ServiceName string `json:"service_name"`
	Environment string `json:"environment"`
	Team        string `json:"team"`
	Notes       string `json:"notes"`
}

const (
	DescribeInstanceCommand string = "describe-instance"
	DeleteVMCommand         string = "delete-vm"
)

func main() {

	var filename = "delete-vm-consumer-risk"
	command := DeleteVMCommand

	RunCommands(command, filename)
}

func RunCommands(command, filename string) {
	records := readCsvFile(fmt.Sprintf("files/%s.csv", filename))
	totalWorker := 5
	totalTask := len(records)

	start := time.Now()

	switch command {
	case DescribeInstanceCommand:
		RunDescribeInstanceCommand(filename, records, totalWorker)
	case DeleteVMCommand:
		RunDeleteVMCommand(filename, records, totalWorker)
	default:
		fmt.Printf("Unknown Command %s\n", command)
	}

	end := time.Now()

	fmt.Printf("----- Execution -----\nstart on: %v\nend on: %v\nTask: %d\nWorker : %d\nduration: %v\n------------------------\n", start, end, totalTask, totalWorker, time.Since(start))

}

func RunDescribeInstanceCommand(filename string, records [][]string, totalWorker int) {
	var instances []Instances

	// Start Worker Pool.
	wp := workerpool.NewWorkerPool(totalWorker)
	wp.Run()

	totalTask := len(records)
	resultC := make(chan Instances, totalTask)

	for i, d := range records {
		d := d
		i := i

		wp.AddTask(

			func() {

				insName := d[0]
				env := d[1]
				team := d[2]

				fmt.Printf("---- Processing Task : %d--- %s -> %s ----- \n", i, insName, env)

				if i > 0 {
					if env == "production" || env == "integration" {
						args := GetDescribeInstanceCommand(insName, env)
						msg := ExecuteCommand(args)

						resultC <- EvaluateOutput(msg, insName, env, team)
					} else {
						resultC <- Instances{
							ServiceName: insName,
							Environment: env,
							Team:        team,
						}
					}
				} else {
					resultC <- Instances{}
				}

			},
		)

	}

	for i := 0; i < totalTask; i++ {
		res := <-resultC
		if i == 0 {
			continue
		}
		instances = append(instances, res)
	}

	CreateCSV(instances, filename, DescribeInstanceCommand)

}

func RunDeleteVMCommand(filename string, records [][]string, totalWorker int) {

	// Start Worker Pool.
	wp := workerpool.NewWorkerPool(totalWorker)
	wp.Run()

	var deleteVM []DeleteVM

	totalTask := len(records)
	resultC := make(chan DeleteVM, totalTask)

	for i, d := range records {
		d := d
		i := i

		wp.AddTask(

			func() {

				insName := d[0]
				env := d[1]
				team := d[2]

				fmt.Printf("---- Processing Task : %d--- %s -> %s ----- \n", i, insName, env)

				if i > 0 {
					if env == "production" || env == "integration" {
						args := GetDeleteVMCommand(insName, env)
						msg := ExecuteCommand(args)

						resultC <- EvaluateDeleteVMOutput(msg, insName, env, team)
					} else {
						resultC <- DeleteVM{
							ServiceName: insName,
							Environment: env,
							Team:        team,
							Notes:       "No Action",
						}
					}
				} else {
					resultC <- DeleteVM{}
				}

			},
		)

	}

	for i := 0; i < totalTask; i++ {
		res := <-resultC
		if i == 0 {
			continue
		}
		deleteVM = append(deleteVM, res)
	}

	CreateCSV(deleteVM, filename, DeleteVMCommand)

}

func GetDescribeInstanceCommand(instanceName, environment string) []string {
	args := []string{
		"execute",
		"describe-instance",
		fmt.Sprintf("INSTANCE_NAME=%s", instanceName),
		fmt.Sprintf("ENVIRONMENT=%s", environment),
	}

	return args

}

func GetDeleteVMCommand(vmName, environment string) []string {

	/*
		proctor execute delete-vm
		VM_NAME=g-fraud-device-tigergraph-db-c-02
		PROJECT_NAME=rabbit-hole-integration-007 PURGE=true DELETE_DISKS=all

	*/

	projectName := "rabbit-hole-integration-007"
	if environment == "production" {
		projectName = "infrastructure-904"
	}

	args := []string{
		"execute",
		"delete-vm",
		fmt.Sprintf("VM_NAME=%s", vmName),
		fmt.Sprintf("PROJECT_NAME=%s", projectName),
		fmt.Sprintf("PURGE=%v", true),
		fmt.Sprintf("DELETE_DISK=%s", "all"),
	}

	return args

}

func ExecuteCommand(args []string) string {

	msg, err := exec.Command("proctor", args...).Output()
	if err != nil {
		log.Println(err)
	}

	return string(msg)
}

func EvaluateDeleteVMOutput(msg, insName, environment, team string) DeleteVM {

	notes := "No Action"

	m := strings.ToLower(msg)
	resps := strings.Split(m, "\n")

	for _, d := range resps {
		if d == "" {
			continue
		}

		deleted := strings.Contains(d, "deleted")
		if deleted {
			notes = d
		}

	}

	return DeleteVM{
		ServiceName: insName,
		Environment: environment,
		Team:        team,
		Notes:       notes,
	}

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

	disk1 := tables["disk1 size"]
	disk2 := tables["disk2 size"]

	if disk1 != "" {
		arr := strings.Split(disk1, " ")
		if len(arr) > 2 {
			disk1 = arr[1]
		}
	}

	if disk2 != "" {
		arr := strings.Split(disk2, " ")
		if len(arr) > 2 {
			disk2 = arr[1]
		}
	}

	diskName1 := tables["disk1 name"]
	diskName2 := tables["disk2 name"]

	return Instances{
		ServiceName: insName,
		IP:          tables["ip"],
		Environment: environment,
		MachineType: tables["machine type"],
		DiskName1:   diskName1,
		Disk1Size1:  disk1,
		DiskName2:   diskName2,
		Disk1Size2:  disk2,
		Team:        team,
	}

}

func CreateCSV(data interface{}, filename, command string) {

	var rows [][]string

	file, err := os.Create(fmt.Sprintf("files/%s-output-%v.csv", filename, time.Now()))
	if err != nil {
		log.Println("Cannot create CSV file:", err)
	}
	defer file.Close()

	switch command {
	case DescribeInstanceCommand:
		rows = append(rows, []string{
			"ServiceName",
			"IP",
			"Environment",
			"MachineType",
			"DiskName1",
			"DiskSize1 GB",
			"DiskName2",
			"DiskSize2 GB",
			"Team",
		})

		ins := data.([]Instances)

		for _, record := range ins {
			row := []string{
				record.ServiceName,
				record.IP,
				record.Environment,
				record.MachineType,
				record.DiskName1,
				record.Disk1Size1,
				record.DiskName2,
				record.Disk1Size2,
				record.Team,
			}
			rows = append(rows, row)
		}
	case DeleteVMCommand:
		rows = append(rows, []string{
			"ServiceName",
			"Environment",
			"Team",
			"Notes",
		})

		ins := data.([]DeleteVM)

		for _, record := range ins {
			row := []string{
				record.ServiceName,
				record.Environment,
				record.Team,
				record.Notes,
			}
			rows = append(rows, row)
		}
	default:
		return
	}

	if len(rows) < 1 {
		return
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
