package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fhs/go-netcdf/netcdf"
)

const (
	// Retrospective is the AWS bucket for retrieving historic (reanalysis) data
	Retrospective string = "noaa-nwm-retro-v2.0-pds"

	// ForecastGCP is the GCP bucket for retrieving forecast data
	ForecastGCP string = "national-water-model"
)

// CheckError simple for error handling. Panic is not ideal for exported code though
func CheckError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// FileExists checks if a file exists, returning bool
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a dir exists, returning bool
func DirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// MountS3Bucket will mount a s3 bucket locally using goofys
func MountS3Bucket(bucketName string, mountLocation string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("goofys %s %s", bucketName, mountLocation))
	// fmt.Println("Mounting s3 bucket with Goofys...")
	_, err := cmd.Output()
	return err
}

// MountGCPBucket will mount a GCP bucket locally using goofys
func MountGCPBucket(bucketName string, mountLocation string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`goofys --endpoint "https://storage.googleapis.com" %s %s`, bucketName, mountLocation))
	// fmt.Println("Mounting GCP bucket with Goofys...")
	_, err := cmd.Output()
	return err
}

// PrintNetcdfVersion will print the netcdf version
func PrintNetcdfVersion() {
	fmt.Println("netCDF library version:", netcdf.Version())
}
