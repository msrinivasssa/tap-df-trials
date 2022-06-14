package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type PackageRepository []struct {
	Name       string `json:"name"`
	Reason     string `json:"reason"`
	Repository string `json:"repository"`
	Status     string `json:"status"`
	Tag        string `json:"tag"`
	Version    string `json:"version"`
}

func executeCommand(command string) (string, error) {

	commandName := strings.Split(command, " ")[0]
	arguments := strings.Split(command, " ")[1:]
	cmd := exec.Command(commandName, arguments...)
	log.Printf("Command to execute: %s", cmd)
	stdoutStderr, err := cmd.CombinedOutput()
	log.Printf("Output: %s Error: %s", stdoutStderr, err)
	return string(stdoutStderr), err

}

func getCurrentRepoVersion(repo_name string, namespace string) (string, error) {

	packageRepositoryVersion := ""

	var getPackageRepository PackageRepository

	getPackageRepositoryCommand := fmt.Sprintf("tanzu package repository get %s -n %s -o json", repo_name, namespace)

	output, err := executeCommand(getPackageRepositoryCommand)

	if strings.HasPrefix(output, "[") {
		err = json.Unmarshal([]byte(output), &getPackageRepository)
	} else {
		outputArray := strings.SplitN(output, "\n", 2)
		strippedOutput := outputArray[1]
		err = json.Unmarshal([]byte(strippedOutput), &getPackageRepository)
	}

	if err != nil {
		log.Printf("Error seen while getting the Package repository details %s \n", err)
	} else {
		log.Printf("Package repository details: %s \n", getPackageRepository)
		packageRepositoryVersion = getPackageRepository[0].Tag

	}
	return packageRepositoryVersion, err

}

func updatePackageRepository(registry string, repository string, version string, repo_name string, repo_namespace string) {
	registry_url := fmt.Sprintf("%s/%s:%s", registry, repository, version)
	PackageRepositoryUpdatecmd := fmt.Sprintf("tanzu package repository update %s --url %s --namespace %s", repo_name, registry_url, repo_namespace)

	stdoutStderr, err := executeCommand(PackageRepositoryUpdatecmd)
	if err != nil {
		log.Printf("Some error during package repo upgrade %s", err)
	} else {
		fmt.Printf("Result: %s \n", string(stdoutStderr))

	}
}

func upgradeTapInstall(meta_pkg_name string, pkg_version string, pkg_install_namespace string, pkg_install_values_yaml string) error {

	upgradeTapInstallCommand := fmt.Sprintf("tanzu package installed update %s -p tap.tanzu.vmware.com -v %s -n %s -f %s", meta_pkg_name, pkg_version, pkg_install_namespace, pkg_install_values_yaml)
	stdoutStderr, err := executeCommand(upgradeTapInstallCommand)
	if err != nil {
		log.Printf("Some error during package repo upgrade %s", err)
	} else {
		fmt.Printf("Result: %s \n", string(stdoutStderr))

	}
	return err
}

func main() {
	tapUpgradeValuesFile := "tap-upgrade-values.yaml"
	tapUpgradeValuesBytes, err := os.ReadFile(tapUpgradeValuesFile)

	tapUpgradeValues := make(map[string]string)

	if err != nil {
		log.Printf("error %s seen while reading TAP upgrade values file %s", err, tapUpgradeValuesFile)

	} else {
		log.Printf("Read TAP upgrade values file \n %s", tapUpgradeValuesBytes)
	}

	// unmarshall
	err = yaml.Unmarshal(tapUpgradeValuesBytes, &tapUpgradeValues)

	if err != nil {
		log.Printf("error while unmarshalling outerloop config file %s", tapUpgradeValuesFile)
		log.Printf("Error log: %s", err)
	} else {
		log.Printf("unmarshalled file %s", tapUpgradeValues)
	}

	currentInstalledRepoVersion, err := getCurrentRepoVersion(tapUpgradeValues["tap_repo_name"], tapUpgradeValues["tap_repo_namespace"])
	log.Printf("Current Installed Version: %s \n", currentInstalledRepoVersion)
	if currentInstalledRepoVersion == tapUpgradeValues["tap_version"] {
		log.Printf("The installed version is %s and is up to date", currentInstalledRepoVersion)
	} else {
		log.Printf("The system needs to be upgraded to %s", tapUpgradeValues["tap_version"])
		updatePackageRepository(tapUpgradeValues["tap_repo_registry"], tapUpgradeValues["tap_repo_repository"], tapUpgradeValues["tap_version"], tapUpgradeValues["tap_repo_name"], tapUpgradeValues["tap_repo_namespace"])
		err = upgradeTapInstall(tapUpgradeValues["tap_pkg_install_name"], tapUpgradeValues["tap_version"], tapUpgradeValues["tap_install_namespace"], tapUpgradeValues["tap_install_values_filename"])
		if err != nil {
			log.Printf("Some issue with package upgrade. error %s found", err)
		} else {
			log.Printf("Package Update successful")
		}
	}

}
