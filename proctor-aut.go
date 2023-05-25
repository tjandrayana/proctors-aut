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
	ServiceName   string `json:"service_name"`
	Environment   string `json:"environment"`
	Team          string `json:"team"`
	ProjectName   string `json:"project"`
	Notes         string `json:"notes"`
	ExecutionDate string `json:"execution_date"`
}

type AddingGCloudLabel struct {
	ResourceName string `json:"service_name"`
	ResourceType string `json:"resource_type"`
	Environment  string `json:"environment"`
	ProjectID    string `json:"project_id"`
	TeamName     string `json:"team_name"`
	Zone         string `json:"zone"`
	AppName      string `json:"app_name"`

	Notes         string `json:"notes"`
	ExecutionDate string `json:"execution_date"`
}

type ResizeVM struct {
	ServiceName   string `json:"service_name"`
	Environment   string `json:"environment"`
	Team          string `json:"team"`
	ProjectName   string `json:"project"`
	UpdatedVMType string `json:"updated_vm_type"`
	Notes         string `json:"notes"`
	ExecutionDate string `json:"execution_date"`
}

const (
	DescribeInstanceCommand  string = "describe-instance"
	DeleteVMCommand          string = "delete-vm"
	AddingGCloudLabelCommand string = "add-gcloud-resource-labels"
	ResizeVMCommand          string = "resize-vm"
)

func main() {

	var filename = "delete_request_25_05_2023"
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
	case AddingGCloudLabelCommand:
		RunAddingGCloudLabelCommand(filename, records, totalWorker)
	case ResizeVMCommand:
		RunResizeVMCommand(filename, records, totalWorker)
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

func RunAddingGCloudLabelCommand(filename string, records [][]string, totalWorker int) {

	// Start Worker Pool.
	wp := workerpool.NewWorkerPool(totalWorker)
	wp.Run()

	var addingGCloudLabel []AddingGCloudLabel

	totalTask := len(records)
	resultC := make(chan AddingGCloudLabel, totalTask)

	for i, d := range records {
		d := d
		i := i

		wp.AddTask(

			func() {

				resourceName := d[0]
				env := d[1]
				teamName := d[2]
				resourceType := d[3]
				projectID := d[4]
				zone := d[5]
				appName := d[6]

				request := AddingGCloudLabel{
					ResourceName: resourceName,
					ResourceType: resourceType,
					Environment:  env,
					ProjectID:    projectID,
					TeamName:     teamName,
					Zone:         zone,
					AppName:      appName,
				}

				fmt.Printf("---- Processing Task : %d--- %s -> %s ----- \n", i, request.ResourceName, env)

				if i > 0 {
					if env == "production" || env == "integration" {
						args := GetAddingGCloudLabelCommand(request)
						msg := ExecuteCommand(args)

						resultC <- EvaluateAddingGCloudLabelOutput(msg, request)
					} else {
						resultC <- AddingGCloudLabel{
							ResourceName: request.ResourceName,
							ResourceType: request.ResourceType,
							Environment:  request.Environment,
							ProjectID:    request.ProjectID,
							TeamName:     request.TeamName,
							Zone:         request.Zone,
							AppName:      request.AppName,

							Notes:         "No Action",
							ExecutionDate: fmt.Sprintf("%d-%d-%d", time.Now().Year(), time.Now().Month(), time.Now().Day()),
						}
					}
				} else {
					resultC <- AddingGCloudLabel{}
				}

			},
		)

	}

	for i := 0; i < totalTask; i++ {
		res := <-resultC
		if i == 0 {
			continue
		}
		addingGCloudLabel = append(addingGCloudLabel, res)
	}

	CreateCSV(addingGCloudLabel, filename, AddingGCloudLabelCommand)

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
				projectName := d[3]

				fmt.Printf("---- Processing Task : %d--- %s -> %s ----- \n", i, insName, env)

				request := DeleteVM{
					ServiceName:   insName,
					Environment:   env,
					Team:          team,
					ProjectName:   projectName,
					ExecutionDate: fmt.Sprintf("%d-%d-%d", time.Now().Year(), time.Now().Month(), time.Now().Day()),
				}

				if i > 0 {
					if env == "production" || env == "integration" {
						args := request.GetDeleteVMCommand()
						msg := ExecuteCommand(args)

						resultC <- request.EvaluateDeleteVMOutput(msg)
					} else {
						request.Notes = "No Action"
						resultC <- request
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

func (d DeleteVM) GetDeleteVMCommand() []string {

	args := []string{
		"execute",
		"delete-vm",
		fmt.Sprintf("VM_NAME=%s", d.ServiceName),
		fmt.Sprintf("PROJECT_NAME=%s", d.ProjectName),
		fmt.Sprintf("PURGE=%v", true),
		fmt.Sprintf("DELETE_DISK=%s", "all"),
	}

	return args

}

func GetAddingGCloudLabelCommand(request AddingGCloudLabel) []string {

	args := []string{
		"execute",
		"add-gcloud-resource-labels",
		fmt.Sprintf("RESOURCES_NAME=%s", request.ResourceName),
		fmt.Sprintf("RESOURCE_TYPE=%s", request.ResourceType),
		fmt.Sprintf("PROJECT_ID=%s", request.ProjectID),
		fmt.Sprintf("TEAM_NAME=%s", request.TeamName),
		fmt.Sprintf("ZONE=%s", request.Zone),
		fmt.Sprintf("APP_NAME=%s", request.AppName),
		fmt.Sprintf("APP=%s", request.AppName),
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

func (d DeleteVM) EvaluateDeleteVMOutput(msg string) DeleteVM {

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

	d.Notes = notes

	return d
}

func EvaluateAddingGCloudLabelOutput(msg string, request AddingGCloudLabel) AddingGCloudLabel {

	notes := "No Action"

	m := strings.ToLower(msg)
	resps := strings.Split(m, "\n")

	fmt.Printf("[EvaluateAddingGCloudLabelOutput] -> Messages %s\n\n\n", m)

	for _, d := range resps {
		if d == "" {
			continue
		}
		fmt.Printf("[EvaluateAddingGCloudLabelOutput] -> %s\n", d)
		note := strings.Contains(d, "added following labels on instance")
		if note {
			notes = d
		}

	}

	return AddingGCloudLabel{
		ResourceName: request.ResourceName,
		ResourceType: request.ResourceType,
		Environment:  request.Environment,
		ProjectID:    request.ProjectID,
		TeamName:     request.TeamName,
		Zone:         request.Zone,
		AppName:      request.AppName,

		Notes:         notes,
		ExecutionDate: fmt.Sprintf("%d-%d-%d", time.Now().Year(), time.Now().Month(), time.Now().Day()),
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
			"Project",
			"Notes",
			"Execution Date",
		})

		ins := data.([]DeleteVM)

		for _, record := range ins {
			row := []string{
				record.ServiceName,
				record.Environment,
				record.Team,
				record.ProjectName,
				record.Notes,
				record.ExecutionDate,
			}
			rows = append(rows, row)
		}

	case AddingGCloudLabelCommand:
		rows = append(rows, []string{
			"ResourceName",
			"Environment",
			"Team",
			"Zone",
			"ResourceType",
			"ProjectID",
			"AppName",
			"Notes",
			"Execution Date",
		})

		ins := data.([]AddingGCloudLabel)

		for _, record := range ins {
			row := []string{
				record.ResourceName,
				record.Environment,
				record.TeamName,
				record.Zone,
				record.ResourceType,
				record.ProjectID,
				record.AppName,
				record.Notes,
				record.ExecutionDate,
			}
			rows = append(rows, row)
		}

	case ResizeVMCommand:

		rows = append(rows, []string{
			"ServiceName",
			"Environment",
			"Team",
			"Project",
			"UpdatedVMType",
			"Notes",
			"Execution Date",
		})

		ins := data.([]ResizeVM)

		for _, record := range ins {
			row := []string{
				record.ServiceName,
				record.Environment,
				record.Team,
				record.ProjectName,
				record.UpdatedVMType,
				record.Notes,
				record.ExecutionDate,
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

func RunResizeVMCommand(filename string, records [][]string, totalWorker int) {

	// Start Worker Pool.
	wp := workerpool.NewWorkerPool(totalWorker)
	wp.Run()

	var resizeVM []ResizeVM

	totalTask := len(records)
	resultC := make(chan ResizeVM, totalTask)

	for i, d := range records {
		d := d
		i := i

		wp.AddTask(

			func() {

				insName := d[0]
				env := d[1]
				team := d[2]
				projectName := d[3]
				updatedVMType := d[4]

				fmt.Printf("---- Processing Task : %d--- %s -> %s ----- \n", i, insName, env)

				request := ResizeVM{
					ServiceName:   insName,
					Environment:   env,
					Team:          team,
					ProjectName:   projectName,
					UpdatedVMType: updatedVMType,
					ExecutionDate: fmt.Sprintf("%d-%d-%d", time.Now().Year(), time.Now().Month(), time.Now().Day()),
				}

				if i > 0 {
					if env == "production" || env == "integration" {
						args := request.GetResizeVMCommand()
						msg := ExecuteCommand(args)

						resultC <- request.EvaluateResizeVMOutput(msg)
					} else {
						request.Notes = "No Action"
						resultC <- request
					}
				} else {
					resultC <- ResizeVM{}
				}

			},
		)

	}

	for i := 0; i < totalTask; i++ {
		res := <-resultC
		if i == 0 {
			continue
		}
		resizeVM = append(resizeVM, res)
	}

	CreateCSV(resizeVM, filename, DeleteVMCommand)

}

func (d ResizeVM) GetResizeVMCommand() []string {

	args := []string{
		"execute",
		"resize-vm",
		fmt.Sprintf("VM_NAME=%s", d.ServiceName),
		fmt.Sprintf("PROJECT_ID=%s", d.ProjectName),
		fmt.Sprintf("UPDATED_VM_TYPE=%v", d.UpdatedVMType),
	}

	return args

}

func (r ResizeVM) EvaluateResizeVMOutput(msg string) ResizeVM {

	notes := "No Action"

	m := strings.ToLower(msg)
	resps := strings.Split(m, "\n")

	for _, d := range resps {
		if d == "" {
			continue
		}

		info := strings.Contains(d, "[info] Action:")
		if info {
			notes = d
		} else {
			notes = fmt.Sprintf("Update %s to be %s in the project %s success", r.ServiceName, r.UpdatedVMType, r.Environment)
		}

	}

	r.Notes = notes

	return r
}
