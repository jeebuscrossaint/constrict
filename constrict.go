package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"bufio"
	"strings"
	"os/exec"
	"runtime"
)

func locate_constrict(filename string) bool {
	files, err := ioutil.ReadDir(".") // read current dir
	if err != nil {
		log.Fatal(err)
		fmt.Println("Error reading directory. Ensure you have read permissions.")
	}

	for _, file := range files {
		if file.Name() == filename {
			return true
		}
	}
	
	return false
}

func parse_vars(filename string) (map[string]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    vars := make(map[string]string)
    scanner := bufio.NewScanner(file)
    lineNumber := 0
    for scanner.Scan() {
        lineNumber++
        line := scanner.Text()
        parts := strings.Split(line, "=")
        if len(parts) != 2 {
            continue
        }
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        // Check if the key already exists
        if _, exists := vars[key]; exists {
            return nil, fmt.Errorf("variable %s is assigned more than once at line %d", key, lineNumber)
        }

        // Check if the value starts with another variable name
        if len(value) > 0 && value[0] == '!' {
            otherVar := value[1:]
            otherValue, exists := vars[otherVar]
            if !exists {
                return nil, fmt.Errorf("variable %s is not defined at line %d", otherVar, lineNumber)
            }
            value = otherValue
        }

        vars[key] = value
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return vars, nil
} 


func parse_functions(filename string) (map[string][]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    funcs := make(map[string][]string)
    scanner := bufio.NewScanner(file)
    var currentFunc string
    for scanner.Scan() {
        line := scanner.Text()
        if line == strings.ToUpper(line) && len(line) > 0 && line[len(line)-1] != '=' {
            // Start of a new function
            if _, exists := funcs[line]; exists {
                return nil, fmt.Errorf("function %s is defined more than once", line)
            }
            currentFunc = line
            funcs[currentFunc] = []string{}
        } else if currentFunc != "" {
            // Part of the current function
            funcs[currentFunc] = append(funcs[currentFunc], line)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return funcs, nil
}

func run_functions(funcs map[string][]string, vars map[string]string) error {
	shell := "/bin/sh"
	flag := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	}

	for _, commands := range funcs {
		for _, command := range commands {
			// replace variable references with values 
			for var_name, var_value := range vars {
				command = strings.Replace(command, "$("+var_name+")", var_value, -1)
			}
			fmt.Println(command) // print the command BEFORE its run.
			cmd := exec.Command(shell, flag, command)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func possible_incremental_building() {

}

func detect_stuff() {

}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    file_exists := locate_constrict("confile")
    if file_exists {
        // fmt.Println("file found")
        vars, err := parse_vars("confile")
        if err != nil {
            log.Fatal(err)
        }
        // fmt.Println(vars)

        funcs, err := parse_functions("confile")
        if err != nil {
            log.Fatal(err)
        }
        // fmt.Println(funcs)

        // Determine which function to run
        funcToRun := "DEFAULT"
        if len(os.Args) > 1 {
            funcToRun = strings.ToUpper(os.Args[1])
        }

        // Check if the function exists
        if commands, exists := funcs[funcToRun]; exists {
            err = run_functions(map[string][]string{funcToRun: commands}, vars)
            if err != nil {
                log.Fatal(err)
            }
        } else {
            fmt.Printf("Function %s not found\n", funcToRun)
        }
    } else {
        fmt.Println("file not found")
	fmt.Println("make one with `touch confile` and add your commands to it")
    }
}

