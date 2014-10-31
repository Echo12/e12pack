package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"errors"
)

var PACK_WORKER = 10

var PACKMETAFILE = ".e12pack"
var PACKSETTINGSFILE = ".e12pack_settings"

type PackerFunc func(config PackerConfigPacker) (PackStrategyFunc, error)
type PackStrategyFunc func(job PackJob) error

var PACKERS = map[string]PackerFunc{
	"pbomanager": PackerPBOManager,
}

func PackerPBOManager(cfg PackerConfigPacker) (PackStrategyFunc, error) {
	_, err := os.Stat(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("Could not find packer executable: %s, error: %s", cfg.Path, err)
	}
	return PackStrategyFunc(func(job PackJob) error {
		cmd := exec.Command(cfg.Path, "-pack", job.packDir, job.outputFile)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error while working on %s: %s", job.packDir, err)
		}
		return nil
	}), nil
}

func main() {
	packPath := flag.String("pack", "", "Path to pack")
	flag.Parse()

	if packPath == nil || *packPath == "" {
		fmt.Println("Packpath missing")
		os.Exit(1)
	}

	if err := pack(*packPath); err != nil {
		fmt.Printf("Error while packing: %s", err)
		os.Exit(1)
	}
}

func pack(packPath string) error {
	packDirInfo, err := os.Stat(packPath)
	if err != nil {
		return err
	}

	if !packDirInfo.IsDir() {
		return fmt.Errorf("Packpath is not a directory: %s", packPath)
	}

	metaFilePath := filepath.Join(packPath, PACKMETAFILE)
	packFileNames, err := readMetaFile(metaFilePath)
	if err != nil {
		return fmt.Errorf("Error while reading meta file: %s %s", metaFilePath, err)
	}

	settingsFilePath := filepath.Join(packPath, PACKSETTINGSFILE)
	packSettings, err := readSettingsFile(settingsFilePath)
	if err != nil {
		return fmt.Errorf("Error while reading settings file: %s %s", settingsFilePath, err)
	}

	outputPath := packSettings.Output
	outputDirInfo, err := os.Stat(outputPath)
	if err != nil {
		return err
	}

	if !outputDirInfo.IsDir() {
		return fmt.Errorf("Outputpath is not a directory: %s", outputPath)
	}

	packerFunc, ok := PACKERS[strings.ToLower(packSettings.PackerConf.Name)]
	if !ok {
		return fmt.Errorf("Error no packer strategy for packer: %s", packSettings.PackerConf.Name)
	}

	packStrategy, err := packerFunc(packSettings.PackerConf)
	if err != nil {
		return err
	}

	jobCh := make(chan PackJob)
	closeCh := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(PACK_WORKER)
	defer func() {
		close(closeCh)
	}()

	for i := 0; i < PACK_WORKER; i++ {
		go func() {
			packWorker(jobCh, closeCh, packStrategy)
			wg.Done()
		}()
	}

	for _, packFileName := range packFileNames {
		path := filepath.Join(packPath, packFileName)
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("Directory from packfile %s is not a directory: %s", packFileName, path)
		}
		jobCh <- PackJob{
			packDir:    path,
			outputFile: filepath.Join(outputPath, strings.ToLower(packFileName)+".pbo"),
		}

	}
	close(jobCh)
	wg.Wait()
	return nil
}

type PackJob struct {
	packDir    string
	outputFile string
}

func packWorker(jobCh chan PackJob, closeCh chan struct{}, packStrategy PackStrategyFunc) {
	for {
		select {
		case job, ok := <-jobCh:
			if !ok {
				return
			}
			fmt.Printf("Worker do: %s to %s\n", job.packDir, job.outputFile)
			_, err := os.Stat(job.outputFile)
			if err == nil {
				fmt.Printf("Remove %s before creating new pbo\n", job.outputFile)
				os.Remove(job.outputFile)
			}
			packStrategy(job)
		case <-closeCh:
			return
		}
	}
}

func readMetaFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		err := errors.New(".e12pack file is empty")
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return lines, nil
}

type PackerConfig struct {
	Output     string             `toml:"output"`
	Rapify     *bool              `toml:"rapify"`
	PackerConf PackerConfigPacker `toml:"packer"`
}

type PackerConfigPacker struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
}

func readSettingsFile(name string) (*PackerConfig, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var conf PackerConfig
	_, err = toml.Decode(string(b), &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil

}
