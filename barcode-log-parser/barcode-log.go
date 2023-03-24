package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const SUMMARY = "SUMMARY"
const CSV_TITLES_1 = "Request\t\t\t\t\t\t\t\tStatus\t\t\tTiming\t\t\t\t\t\tMetadata\t\t\t\t"
const CSV_TITLES_2 = "Ts\tOrigen\tFormato\tInfo\tProvider \tTail\tImgLength\tID\tStatus\tStatusCode\tStatusDesc\tTotal\tDecoding\tReading\tH1\tH2\tH3\tHeurDetected \tGeometry\tRegion\tRatio\tConfidence\tContextInfo"
const fechaFormatIn = "02/01/06 15:04:05.000"
const fechaFormatOut = "2006-01-02T15:04:05.000"

type Summary struct {
	Request struct {
		Ts        string `json:"ts"`
		Origen    string `json:"origen"`
		Formato   string `json:"formato"`
		Info      string `json:"info"`
		Provider  string `json:"provider"`
		Tail      string `json:"tail"`
		ImgLength int    `json:"imgLength"`
		ID        string `json:"id"`
	} `json:"request"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	StatusDesc string `json:"statusDesc"`
	Timing     struct {
		Total      int `json:"total"`
		Decoding   int `json:"decoding"`
		Reading    int `json:"reading"`
		Heuristics struct {
			H1 int `json:"h1"`
			H2 int `json:"h2"`
			H3 int `json:"h3"`
		} `json:"heuristics"`
	} `json:"timing"`
	Metadata struct {
		HeurDetected string `json:"heurDetected"`
		Geometry     string `json:"geometry"`
		Region       string `json:"region"`
		Ratio        string `json:"ratio"`
		Confidence   int    `json:"confidence"`
	} `json:"metadata"`
	ContextInfo string `json:"contextInfo"`
}

/////////////////////  FUNCIONES

func help(exit bool) {
	fmt.Println("")
	fmt.Println("barcode-log:  parseo de log barcode")
	fmt.Println("")
	fmt.Println("Uso: ")
	go fmt.Println("")
	fmt.Println(" barcode-log <file> [formato:JSON.default|CSV]")
	fmt.Println("")
	fmt.Println("")

	os.Exit(1)
}

// Va leyendo linea x linea, pero solo devuelve lineas summary
func readSummaryLine(r *bufio.Reader) (string, error) {
	var (
		isSummary bool  = false
		err       error = nil
		line      []byte
		cadena    string
	)

	for !isSummary && err == nil {
		line, _, err = r.ReadLine()
		cadena = string(line)
		isSummary = strings.Contains(cadena, SUMMARY)
	}

	summLine := string(line)
	if len(summLine) > 1 {

		// esto es para las versiones anteriores que no tenian el TS, lo saco del log y lo agrego dentro del request
		if !strings.Contains(summLine, "\"ts\"") {
			summLine = strings.Replace(summLine, "\"origen\":",
				fmt.Sprintf("\"ts\": \"%s\" , \"origen\":", getFechaFromLog(summLine)), 1)
		}
	}

	return summLine, err
}

func getFechaFromLog(buff string) string {
	fechaStr := buff[:21]
	fechaStr = strings.Replace(fechaStr, ",", ".", 1)

	fecha, err := time.Parse(fechaFormatIn, fechaStr)
	if err != nil {
		return "error.parse"
	}

	return fecha.Format(fechaFormatOut)
}

// toma una linea summary y devuelve solo el json
func getJsonText(buff string) []byte {
	inicio := strings.Index(buff, SUMMARY)
	retVal := buff[inicio+len(SUMMARY)+1:]
	return []byte(retVal)

}

// genera salida csv
func writeCSV(summ Summary) {

	req := fmt.Sprintf(" %-25s\t %-30s\t %-8s \t %-8s \t %-20s \t %s \t %8d\t %s \t ",
		summ.Request.Ts,
		summ.Request.Origen,
		summ.Request.Formato,
		summ.Request.Info,
		summ.Request.Provider,
		summ.Request.Tail,
		summ.Request.ImgLength,
		summ.Request.ID)

	status := fmt.Sprintf(" %5s \t %d \t %-80s \t ",
		summ.Status,
		summ.StatusCode,
		summ.StatusDesc)

	timing := fmt.Sprintf(" %5d \t %5d \t  %5d \t %5d \t  %5d \t %5d \t",
		summ.Timing.Total,
		summ.Timing.Decoding,
		summ.Timing.Reading,
		summ.Timing.Heuristics.H1,
		summ.Timing.Heuristics.H2,
		summ.Timing.Heuristics.H3,
	)

	metadata := fmt.Sprintf(" %2s \t %8s \t%8s \t %12s \t %2d \t %s \t",
		summ.Metadata.HeurDetected,
		summ.Metadata.Geometry,
		summ.Metadata.Region,
		summ.Metadata.Ratio,
		summ.Metadata.Confidence,
		summ.ContextInfo)

	fmt.Printf("%s %s %s %s \n", req, status, timing, metadata)

}

func procesa(fileIN, formatoOUT string) {
	primerReg := true

	f, err := os.Open(fileIN)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		os.Exit(1)
	}

	///////  PRE-IMPRESION
	if formatoOUT == "JSON" {
		fmt.Println("[")
	} else {
		fmt.Println(CSV_TITLES_1)
		fmt.Println(CSV_TITLES_2)
	}

	///////  PROCESAMIENTO

	r := bufio.NewReader(f)
	summaryLine, e := readSummaryLine(r)
	for e == nil {

		summaryAsJson := getJsonText(summaryLine)

		if formatoOUT == "CSV" {
			var summ Summary
			err := json.Unmarshal(summaryAsJson, &summ)
			if err != nil {
				fmt.Println(err)
			}

			writeCSV(summ)
		} else {
			if !primerReg {
				fmt.Println(" ,")

			}
			fmt.Println(string(summaryAsJson))
		}

		summaryLine, e = readSummaryLine(r)
		primerReg = false
	}

	///////  POST-IMPRESION

	if formatoOUT == "JSON" {
		fmt.Println("]")
	}
}

/////////////////////  PRINCIPAL

func main() {
	formatoOUT := "JSON"

	if len(os.Args) == 1 {
		help(true)
	}

	if len(os.Args) == 3 {
		formatoOUT = os.Args[2]
	}

	procesa(os.Args[1], formatoOUT)

}
