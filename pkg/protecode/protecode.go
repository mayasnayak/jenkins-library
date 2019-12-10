package protecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

const DELIMITER = "-DeLiMiTeR-"

type ProteCodeProductData struct {
	Products []ProteCodeProduct `json:"products"`
}

type ProteCodeProduct struct {
	ProductId string `json:"product_id"`
}

type ProteCodeResultData struct {
	Result ProteCodeResult `json:"results"`
}

type ProteCodeResult struct {
	ProductId  string               `json:"product_id"`
	ReportUrl  string               `json:"report_url`
	Status     string               `json:"status"`
	Components []ProteCodeComponent `json:"components,omitempty"`
}

type ProteCodeComponent struct {
	Vulnerability []ProteCodeVulnerability `json:"vulns,omitempty"`
}

type ProteCodeVulnerability struct {
	Exact  bool          `json:"exact"`
	Vuln   ProteCodeVuln `json:"vuln"`
	Triage string        `json:"triage"`
}

type ProteCodeVuln struct {
	Cve        string  `json:"cve"`
	Cvss       float64 `json:"cvss"`
	Cvss3Score string  `json:"cvss3_score"`
}

func CreateUrl(pURL string, path string, pValue string, fParam string) *url.URL {

	// Let's start with a base url
	protecodeUrl, err := url.Parse(pURL)
	if err != nil {
		log.Entry().WithError(err).Fatal("Malformed URL")
		os.Exit(1)
	}

	// Add a Path Segment (Path segment is automatically escaped)
	if len(path) > 0 {
		protecodeUrl.Path += fmt.Sprintf("%v", path)
	}

	// Add a Path Segment (Path segment is automatically escaped)
	if len(pValue) > 0 {
		protecodeUrl.Path += fmt.Sprintf("%v", pValue)
	}

	// Prepare Query Parameters
	if len(fParam) > 0 {
		encodedFParam := url.QueryEscape(fParam)
		params := url.Values{}
		params.Add("q", fmt.Sprintf("file:%v", encodedFParam))

		// Add Query Parameters to the URL
		protecodeUrl.RawQuery = params.Encode() // Escape Query Parameters
	}

	return protecodeUrl
}

func CreateRequestHeader(credentialsId string, verbose bool, customHeaders map[string][]string) map[string][]string {
	headers := map[string][]string{
		"authentication":         []string{credentialsId},
		"quiet":                  []string{fmt.Sprintf("%v", !verbose)},
		"ignoreSslErrors":        []string{"true"},
		"consoleLogResponseBody": []string{fmt.Sprintf("%v", verbose)},
	}
	for k, p := range customHeaders {
		headers[k] = p
	}

	return headers
}

func GetProteCodeResultData(r *http.Response) *ProteCodeResultData {
	defer r.Body.Close()

	response := new(ProteCodeResultData)

	err := json.NewDecoder(r.Body).Decode(response)

	if err != nil {
		log.Entry().WithError(err).Fatalf("error during decode response: %v", r.Body)
		//TODO check if this is needed
		os.Exit(1)
	}

	return response
}

func CmdExecGetProtecodeResult(cmdName string, cmdString string) ProteCodeResult {

	var response ProteCodeResult = ProteCodeResult{}
	c := command.Command{}
	c.Dir(".")

	buf := new(bytes.Buffer)
	c.Stdout(buf)

	script := fmt.Sprintf("%v %v", cmdName, cmdString)

	err := command.RunShell("/bin/bash", script)
	if err != nil {
		log.Entry().WithError(err).Fatalf("Failed to exec cmd %v: %v ", cmdName, cmdString)
	}

	parts := strings.Split(buf.String(), DELIMITER)

	if len(parts) > 1 && parts[1] != "status=200" || len(parts) <= 1 {
		log.Entry().WithError(err).Fatalf("Failed to exec cmd %v: %v ", cmdName, cmdString)
		os.Exit(1)
	}

	if err := json.Unmarshal([]byte(parts[0]), response); err != nil {
		log.Entry().WithError(err)
	}

	return response
}

func SendApiRequest(methode string, url string, headers map[string][]string, client Client) *http.Response {

	r, err := client.SendRequest(methode, url, nil, headers, nil)
	if err != nil {
		log.Entry().WithError(err).Fatalf("error during %v: %v reuqest", methode, url)
		os.Exit(1)
	}

	return r
}

func ParseResultToInflux(result ProteCodeResult, protecodeExcludeCVEs string) map[string]int {
	var m map[string]int = make(map[string]int)
	m["count"] = 0
	m["cvss2GreaterOrEqualSeven"] = 0
	m["cvss3GreaterOrEqualSeven"] = 0
	m["historical_vulnerabilities"] = 0
	m["triaged_vulnerabilities"] = 0
	m["excluded_vulnerabilities"] = 0
	m["minor_vulnerabilities"] = 0
	m["major_vulnerabilities"] = 0
	m["vulnerabilities"] = 0

	for _, components := range result.Components {
		for _, vulnerability := range components.Vulnerability {
			if vulnerability.Exact {
				if isExcluded(vulnerability, protecodeExcludeCVEs) {
					m["excluded_vulnerabilities"]++
				} else if isTriaged(vulnerability) {
					m["triaged_vulnerabilities"]++
				} else {
					m["count"]++
					if isSevereCVSS3(vulnerability) {
						m["cvss3GreaterOrEqualSeven"]++
						m["major_vulnerabilities"]++
					} else if isSevereCVSS2(vulnerability) {
						m["cvss2GreaterOrEqualSeven"]++
						m["major_vulnerabilities"]++
					} else {
						m["minor_vulnerabilities"]++
					}
					m["vulnerabilities"]++
				}
			} else {
				m["historical_vulnerabilities"]++
			}
		}
	}

	return m
}

func isExcluded(vulnerability ProteCodeVulnerability, protecodeExcludeCVEs string) bool {
	return strings.Contains(protecodeExcludeCVEs, vulnerability.Vuln.Cve)
}

func isTriaged(vulnerability ProteCodeVulnerability) bool {
	return len(vulnerability.Triage) > 0
}

func isSevereCVSS3(vulnerability ProteCodeVulnerability) bool {
	threshold := 7.0
	cvss3, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss3Score, 64)
	return cvss3 >= threshold
}

func isSevereCVSS2(vulnerability ProteCodeVulnerability) bool {
	threshold := 7.0
	cvss3, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss3Score, 64)
	return cvss3 == 0 && vulnerability.Vuln.Cvss >= threshold
}
