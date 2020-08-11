package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fhs/go-netcdf/netcdf"
)

// constants
const (

	// UserDateFormat style for nwm netcdf file names (YYYY-MM-DD-HH)
	userDateFormat string = "2006-01-02-15"

	// Format style for nwm netcdf forecast file names (YYYYMMDD)
	nwmDateFormat string = "20060102"

	// Format style for nwm netcdf retrospectivefile names (YYYYMMDDHH)
	retrospectiveDateFormat string = "200601021504"

	// FirstRetrospective is the first forecast available from the reanalysis
	FirstRetrospective string = "19930101"

	// FirstEnsembleGCP is the first medium range forecast with ensemble members
	FirstEnsembleGCP string = "20190602"

	// FirstSingleGCP is the first day of forecast available on GCP
	FirstSingleGCP string = "20180917"
)

// Paths is a container for passing netcdf locations from the appropriate bucket
type Paths []string

// ComIDs is an array of the NWM output locations
type ComIDs []int64

// StreamFlow is a single NWM forecast data point
type StreamFlow struct {
	Time       string  `json:"forecast_time"`
	Value      float64 `json:"flow_value"`
	Product    string  `json:"forecast_product"`
	ComidIndex uint64  `json:"netcdf_comid_index"`
}

// ComIDPrediction ...
type ComIDPrediction struct {
	Comid int64   `json:"comid"`
	Value float64 `json:"flow"`
}

// // ForecastAtTime ...
// type ForecastAtTime struct {
// 	Time string            `json:"forecast_time"`
// 	Data []ComIDPrediction `json:"forecast_hour"`
// }

// GetForecastSource returns infromation identifying forecast vintage
// used for selecting source bucket and naming conventions
func GetForecastSource(dtm string) (string, error) {

	forecastDTM, _ := time.Parse(userDateFormat, dtm)
	firstRetro, _ := time.Parse(nwmDateFormat, FirstRetrospective)
	gcpAvailable, _ := time.Parse(nwmDateFormat, FirstSingleGCP)
	ensembleAvailable, _ := time.Parse(nwmDateFormat, FirstEnsembleGCP)

	switch {

	case forecastDTM.Before(firstRetro):
		errMessage := fmt.Sprintf("Forecast unavailable prior to %s", FirstRetrospective)
		return "", errors.New(errMessage)

	case forecastDTM.After(time.Now()):
		errMessage := fmt.Sprintf("Forecast Not Yet Available")
		return "", errors.New(errMessage)

	case forecastDTM.Before(gcpAvailable):
		return "Retrospective-Era", nil

	case forecastDTM.Before(ensembleAvailable):
		return "SingleMember-Era", nil

	default:
		return "Ensemble-Era", nil
	}

}

// GetRetrospectivePaths creates a new ...
func GetRetrospectivePaths(dtm string) []string {

	defaultTimeDays := 5

	forecastDTM, _ := time.Parse(userDateFormat, dtm)
	forecastDTMStart := forecastDTM.AddDate(0, 0, -defaultTimeDays)
	forecastDTMEnd := forecastDTM.AddDate(0, 0, defaultTimeDays)

	nForecastHours := forecastDTMEnd.Sub(forecastDTMStart).Hours()
	cloudPaths := make([]string, 0)

	// startYearMonthDay, startHour, FcastHr
	forecastPath := "retrospective/full_physics/%d/%s.CHRTOUT_DOMAIN1.comp"

	for i := 1; i < int(nForecastHours); i++ {

		t := forecastDTMStart.Add(time.Hour * time.Duration(i))
		timeStamp := t.Format("2006010215") + "00" // floor round the minutes down

		path := fmt.Sprintf(forecastPath, t.Year(), timeStamp)

		cloudPaths = append(cloudPaths, path)
	}

	return cloudPaths
}

// GetShortRangePaths creates a new ...
func GetShortRangePaths(dtm string) []string {

	forecastDTM, _ := time.Parse(userDateFormat, dtm)
	cloudPaths := make([]string, 0)

	startYearMonthDay := forecastDTM.Format(nwmDateFormat)
	startHour := fmt.Sprintf("%02d", forecastDTM.Hour())

	// startYearMonthDay, startHour, FcastHr
	forecastPath := "forecast/nwm.%s/short_range/nwm.t%sz.short_range.channel_rt.f%s.conus.nc"

	for i := 1; i < 19; i++ {
		nextTime := fmt.Sprintf("%03d", i)
		path := fmt.Sprintf(forecastPath, startYearMonthDay, startHour, nextTime)

		cloudPaths = append(cloudPaths, path)
	}

	return cloudPaths
}

// GetMediumRangePaths creates a new ...
func GetMediumRangePaths(dtm string, forecastEra string) []string {

	var forecastPath string
	var startYearMonthDay string
	var startHour string
	var nMembers []int

	cloudPaths := make([]string, 0)

	forecastDTM, _ := time.Parse(userDateFormat, dtm)
	for {
		if !(forecastDTM.Hour() == 0 || forecastDTM.Hour() == 6 || forecastDTM.Hour() == 12 || forecastDTM.Hour() == 18) {
			forecastDTM = forecastDTM.Add(time.Hour * -1)
		} else {
			break
		}
	}
	roundedStart := time.Date(forecastDTM.Year(), forecastDTM.Month(), forecastDTM.Day(), 0, 0, 0, 0, forecastDTM.Location())

	startYearMonthDay = forecastDTM.Format(nwmDateFormat)
	requestedHour := forecastDTM.Hour()

	switch forecastEra {
	case "SingleMember-Era":
		// startYearMonthDay, member, startHour, member, FcastHr
		forecastPath = "forecast/nwm.%s/medium_range/nwm.t%sz.medium_range.channel_rt.f%s.conus.nc"
		nMembers = []int{0}

	case "Ensemble-Era":
		// startYearMonthDay, member, startHour, member, FcastHr
		forecastPath = "forecast/nwm.%s/medium_range_mem%d/nwm.t%sz.medium_range.channel_rt_%d.f%s.conus.nc"
		nMembers = []int{1, 2, 3, 4, 5, 6, 7}

	default:
		// startYearMonthDay, member, startHour, member, FcastHr
		forecastPath = "forecast/nwm.%s/medium_range_mem%d/nwm.t%sz.medium_range.channel_rt_%d.f%s.conus.nc"
		nMembers = []int{2}

	}

	// round down to the closest medium range starting hour
	switch {

	case requestedHour < 6:
		startHour = "00"

	case 6 <= requestedHour && requestedHour < 12:
		startHour = "06"

	case 12 <= requestedHour && requestedHour < 18:
		startHour = "12"

	case 18 <= requestedHour && requestedHour <= 23:
		startHour = "18"

	default:
		startHour = "00"
		startYearMonthDay = roundedStart.Add(time.Hour * 24).Format(nwmDateFormat)
	}

	if forecastEra == "SingleMember-Era" {
		// 207 represents the number of hours from start to 8.5 days
		// Member 1 has more time, however for consistency we'll uese
		// 8.5 for all
		for i := 3; i <= 204; i += 3 {
			nextTime := fmt.Sprintf("%03d", i)
			path := fmt.Sprintf(forecastPath, startYearMonthDay, startHour, nextTime)
			cloudPaths = append(cloudPaths, path)
		}
	} else {
		for _, mem := range nMembers {
			for i := 3; i <= 204; i += 3 {
				nextTime := fmt.Sprintf("%03d", i)
				path := fmt.Sprintf(forecastPath, startYearMonthDay, mem, startHour, mem, nextTime)
				cloudPaths = append(cloudPaths, path)
			}

		}

	}
	return cloudPaths
}

// GetNetCDFData ...
func GetNetCDFData(path string, product *string, idxs *[]uint64) ([]StreamFlow, error) {

	results := make([]StreamFlow, len(*idxs))
	var flow float64

	if !FileExists(path) {
		errMessage := fmt.Sprintf("File does not exist, verify path and mount point %s", path)
		return results, errors.New(errMessage)
	}
	nc, err := netcdf.OpenFile(path, netcdf.NOWRITE)
	defer nc.Close()

	if err != nil {
		errMessage := fmt.Sprintf("Error Opening netcdf file: %s", path)
		return results, errors.New(errMessage)
	}

	varName, err := nc.Var("streamflow")
	if err != nil {
		errMessage := fmt.Sprintf("Error accessing netcdf variable: %s", path)
		return results, errors.New(errMessage)
	}

	for i, idx := range *idxs {

		flow, err = varName.ReadFloat64At([]uint64{idx})
		if err != nil {
			// log.Println("Path =", path)
			// log.Println(err)
			// log.Println("Error reading streamflow, setting equal to -9999...")
			flow = -9999 * 100
		}

		if *product == "Retrospective" {
			forecastDTM := ParseRetrospectiveDate(path)
			results[i] = StreamFlow{forecastDTM, flow / 100, *product, idx}

		} else {
			forecastDTM := ParseNWMDate(path)
			results[i] = StreamFlow{forecastDTM, flow / 100, *product, idx}
		}

	}

	return results, nil

}

// ParseNWMDate ...
func ParseNWMDate(path string) string {

	// Examples
	// forecast/nwm.20190102/medium_range/nwm.t06z.medium_range.channel_rt.f003.conus.nc
	// forecast/nwm.20190102/short_range/nwm.t11z.short_range.channel_rt.f001.conus.nc

	pathParts := strings.Split(path, "/")
	startYearMonthDay := strings.Split(pathParts[1], ".")[1]
	forecastIssueHour := strings.Split(pathParts[3], ".")[1]
	forecastHourOffset := strings.Split(pathParts[3], ".")[4]

	yr, err := strconv.Atoi(startYearMonthDay[:4]) // [2019] 0102
	CheckError(err)

	mo, err := strconv.Atoi(startYearMonthDay[4:6]) // 2019 [01] 02
	CheckError(err)

	day, err := strconv.Atoi(startYearMonthDay[6:]) // 201901 [02]
	CheckError(err)

	hr, err := strconv.Atoi(forecastIssueHour[1:3]) // t [11] z
	CheckError(err)

	fHrInt, err := strconv.Atoi(forecastHourOffset[1:]) // f [001]
	CheckError(err)

	dtm := time.Date(yr, time.Month(mo), day, hr, 0, 0, 0, time.UTC).Add(time.Hour * time.Duration(fHrInt))

	return dtm.Format(userDateFormat)

}

// ParseRetrospectiveDate ...
func ParseRetrospectiveDate(path string) string {
	// Examples
	// fetch/retrospective/full_physics/1993/1993 12 31 23 00.LAKEOUT_DOMAIN1.comp

	fileName := filepath.Base(path)
	dtm := strings.Split(fileName, ".")[0]
	forecastDTM, _ := time.Parse(retrospectiveDateFormat, dtm)
	return forecastDTM.Format(userDateFormat)
}

type nextLevel = map[string][]ComIDPrediction

// FinalResults ...
type FinalResults map[string]nextLevel

// MarshalResults ...
func MarshalResults(results [][]StreamFlow) FinalResults {

	finalResults := make(FinalResults, 0)

	for i := 0; i < len(results); i++ {
		for _, result := range results[i] {

			// Add lookup map here to convert index to comid and drop int conversion
			// comid := lookup[result.ComidIndex]
			comid := int64(result.ComidIndex)
			flowValue := result.Value
			validTime := result.Time
			product := result.Product

			// Instatiate/Check top level (product)
			if _, ok := finalResults[product]; !ok {
				bottomLevel := make([]ComIDPrediction, 0)
				addLevel := nextLevel{validTime: bottomLevel}
				finalResults[product] = addLevel
			}

			// Instatiate/Check next and bottom levels  (time)
			if _, ok := finalResults[product][validTime]; !ok {
				bottomLevel := make([]ComIDPrediction, 0)
				finalResults[product][validTime] = bottomLevel

			}

			// Append result where approriate
			newData := ComIDPrediction{comid, flowValue}
			productTimeData := finalResults[product][validTime]
			finalResults[product][validTime] = append(productTimeData, newData)

		}
	}

	return finalResults
}
