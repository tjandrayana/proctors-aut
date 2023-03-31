
# Proctor Automation

This is an automation that will help to run the proctor Gojek commands in parallels.

Currently there are 2 commands that provided : 
1. Describe Instances
2. Delete VM


## Prerequisite

- Proctor Gojek
- VPN Access

## How to use it

- Create a directory named it as "files" inside the proctor-aut
- Put the CSV File named it as "input.csv"
- Set the worker with you need
- Set the command that you want to executed
- Run the service with this command
```
    go run proctor-aut.go
```

## CSV Format
```
    Service Name,Environment,Team
    service-01,integration,driver
```

## Results

```
    ----- Execution -----
start on: 2023-03-10 17:27:03.176213 +0700 WIB m=+0.000184917
end on: 2023-03-10 17:30:16.61314 +0700 WIB m=+193.436891084
Worker : 10
Total Task: 50
duration: 3m13.436706375s
```


## Important !!!

Don't forget to give stars for this repo. Thank you!!!
