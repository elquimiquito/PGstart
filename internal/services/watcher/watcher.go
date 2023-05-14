package watcher

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	Directories []string `yaml:"path"`
	Commands    []string `yaml:"commands"`
}

func Watcher(name string) {
	config := readConfig(name)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for _, dir := range config.Directories {
		err = watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	var cmd *exec.Cmd
	var cmdOutput []byte
	var cmdErr error
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Ignore changes to temporary files
			if strings.HasPrefix(filepath.Base(event.Name), ".") {
				continue
			}
			fmt.Printf("Change detected: %s\n", event.Name)
			// Execute the commands in order
			for _, command := range config.Commands {
				cmd = exec.Command("bash", "-c", command)
				cmd.Dir = filepath.Dir(event.Name)
				cmdOutput, cmdErr = cmd.Output()
				if cmdErr != nil {
					log.Printf("Error executing command: %s\n", cmdErr)
					break
				}
				log.Printf("Command output: %s\n", string(cmdOutput))
				// Store the change history in the database
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching directory: %s\n", err)
		}
	}
}

func readConfig(name string) Config {
	configFile, err := filepath.Abs("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	configData, err := readFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func readFile(filename string) ([]byte, error) {
	data, err := exec.Command("bash", "-c", fmt.Sprintf("cat %s", filename)).Output()
	if err != nil {
		return nil, err
	}
	return data, nil
}
