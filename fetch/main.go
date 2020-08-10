package main

import (
	"encoding/json"
	"fetch/utils"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {

	// Mount S3 Bucket for fetching data from the retrospective archive
	err := utils.MountS3Bucket(utils.Retrospective, "retrospective")
	if err != nil {
		log.Fatal(err)
	}

	// Mount GCP Bucket for fetching data from the forecast archive
	err = utils.MountGCPBucket(utils.ForecastGCP, "forecast")
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	// Flags for Command Line Arguments
	var requestDateTime, forecastProduct string

	// By Default, get the timestamp for the forecast issued 2 hours from the current time
	// Testing indicates this forecast will always be available, while the last hour may not
	lastHour := fmt.Sprintf(time.Now().UTC().Add(-time.Hour * time.Duration(2)).Format("2006-01-02-15"))
	flag.StringVar(&requestDateTime, "date_time", lastHour, "requested publication hour")

	// Index (position) of streamflow data for comid of interest (900 selected at random)
	flag.StringVar(&forecastProduct, "product", "short", "requested product: medium or short")
	idx0 := flag.Int64("comid", 900, "desired comid for streamflow value")
	flag.Parse()

	idxMap, err := utils.PositionCSVToMap("netcdf_index.csv")
	utils.CheckError(err)

	// Add all requested indices to idxs var for processing
	v, ok := idxMap[*idx0]
	if !ok {
		errMessage := fmt.Sprintf("Comid %d is not in the netcdf file", *idx0)
		log.Fatal(errMessage)
	}
	var idxs []uint64 = []uint64{v}
	additionalIDXs := flag.Args()

	if len(additionalIDXs) > 0 {
		for _, idx := range additionalIDXs {
			idxINT, _ := strconv.ParseInt(idx, 10, 64)
			v, ok := idxMap[idxINT]
			if !ok {
				errMessage := fmt.Sprintf("Comid %d is not in the netcdf file", idxINT)
				log.Fatal(errMessage)
			}
			idxs = append(idxs, uint64(v))
		}
	}

	// Prep variables for storing results
	var isForecast bool
	var paths []string
	var availableProducts []string
	var forecastResults [][]utils.StreamFlow
	var nWorkers int

	// Evaluate which source to pull data from
	forecastEra, err := utils.GetForecastSource(requestDateTime)

	if err != nil {
		errMessage := fmt.Sprintf("No data available for this period: %s", requestDateTime)
		// fmt.Println(errMessage)
		log.Fatal(errMessage)
	}

	// Create list of file paths for requested data
	switch {

	case forecastEra == "Retrospective-Era":
		paths = utils.GetRetrospectivePaths(requestDateTime)
		product := "Retrospective"
		availableProducts = utils.AppendIfMissing(availableProducts, product)
		nWorkers = 20
		isForecast = false

	case forecastProduct == "short":
		paths = utils.GetShortRangePaths(requestDateTime)
		nWorkers = 6
		isForecast = true

	case forecastProduct == "medium":
		paths = utils.GetMediumRangePaths(requestDateTime, forecastEra)
		nWorkers = 18
		isForecast = true

	default:
		paths = utils.GetShortRangePaths(requestDateTime)
		mrPaths := utils.GetMediumRangePaths(requestDateTime, forecastEra)
		for _, p := range mrPaths {
			paths = append(paths, p)
		}
		nWorkers = 18
		isForecast = true

	}

	if isForecast {
		// gets available products
		for _, f := range paths {
			product := strings.Split(filepath.Base(f), ".")[2]
			availableProducts = utils.AppendIfMissing(availableProducts, product)
		}
	}

	// create a job queue channel
	jobsChan := make(chan string, len(paths))

	// create a buffered channel [len(paths)] that will be populated with results by the worker function
	resultsChan := make(chan []utils.StreamFlow, len(paths))

	// start up the worker function go routines, they are now waiting for the job queue to be populated
	for i := 0; i < nWorkers; i++ {
		go utils.Worker(jobsChan, resultsChan, &idxs)
	}

	// populate the job queue
	for _, f := range paths {
		jobsChan <- f
	}
	// close the job channel so that no more jobs can be added
	close(jobsChan)

	// pull the results out of the results channel as they are added (by the go routines running the worker function)
	// when the results channel is empty, it will be closed
	// we are controlling for this by limiting the length of our for loop [len(paths)]
	for i := 0; i < len(paths); i++ {
		forecastResults = append(forecastResults, <-resultsChan)
		// Add a check here to verify all results populated
	}

	finalResults := utils.MarshalResults(forecastResults)
	finalResultsBytes, err := json.Marshal(finalResults)
	utils.CheckError(err)
	fmt.Println(string(finalResultsBytes))
}
