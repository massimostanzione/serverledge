package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/grussorusso/serverledge/internal/api"
	"github.com/grussorusso/serverledge/internal/config"
)

var serverConfig config.RemoteServerConf

func exitWithUsage() {
	fmt.Println("expected 'invoke' or 'create' subcommands")
	os.Exit(1)
}

func main() {
	config.ReadConfiguration()

	// Set defaults
	serverConfig.Host = "127.0.0.1"
	serverConfig.Port = config.GetInt("api.port", 1323)

	// Parse general configuration
	flag.IntVar(&serverConfig.Port, "port", serverConfig.Port, "port for remote connection")
	flag.StringVar(&serverConfig.Host, "host", serverConfig.Host, "host for remote connection")
	flag.Parse()

	if len(os.Args) < 2 {
		exitWithUsage()
	}

	switch os.Args[1] {

	case "invoke":
		invoke()
	case "create":
		create()
	default:
		exitWithUsage()
	}
}

type paramsFlags map[string]string

func (i *paramsFlags) String() string {
	return fmt.Sprintf("%q", *i)
}

func (i *paramsFlags) Set(value string) error {
	tokens := strings.Split(value, ":")
	if len(tokens) != 2 {
		return fmt.Errorf("Invalid argument")
	}
	(*i)[tokens[0]] = tokens[1]
	return nil
}

func invoke() {
	var params paramsFlags = make(map[string]string)

	invokeCmd := flag.NewFlagSet("invoke", flag.ExitOnError)
	funcName := invokeCmd.String("function", "", "name of the function")
	invokeCmd.Var(&params, "param", "Function parameter: <name>:<value>")
	invokeCmd.Parse(os.Args[2:])

	if len(*funcName) < 1 {
		fmt.Printf("Invalid function name.\n")
		exitWithUsage()
	}
	invocationBody, err := json.Marshal(params)
	if err != nil {
		exitWithUsage()
	}

	// Send invocation request
	url := fmt.Sprintf("http://%s:%d/invoke/%s", serverConfig.Host, serverConfig.Port, *funcName)
	resp, err := postJson(url, invocationBody)
	if err != nil {
		fmt.Printf("Invocation failed: %v", err)
		os.Exit(2)
	}
	printJsonResponse(resp.Body)
}

func getSourcesTarFile(srcPath string) (*os.File, error) {
	fileInfo, err := os.Stat(srcPath)
	if err != nil {
		return nil, fmt.Errorf("Missing source file")
	}

	if fileInfo.IsDir() || strings.HasSuffix(srcPath, ".tar") {
		// TODO: create a tar archive
	} else {
		// this is a tar file
		// TODO: just return it
	}

	return nil, nil // TODO
}

func create() {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	funcName := createCmd.String("function", "", "name of the function")
	runtime := createCmd.String("runtime", "python38", "runtime for the function")
	handler := createCmd.String("handler", "", "function handler")
	memory := createCmd.Int("memory", 128, "max memory in MB for the function")
	src := createCmd.String("src", "", "source the function (single file, directory or TAR archive)")
	createCmd.Parse(os.Args[2:])

	// TODO: create base64-encoded source Tar
	// 1) Check whether src is a TAR archive, a directory or a generic file
	// 2) TAR archive: just encode
	// 3) file/directory: create TAR and execute step 2)
	_, err := getSourcesTarFile(*src)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(3)
	}

	request := api.FunctionCreationRequest{Name: *funcName, Handler: *handler, Runtime: *runtime, Memory: *memory, SourceTarBase64: *src}
	requestBody, err := json.Marshal(request)
	if err != nil {
		exitWithUsage()
	}

	url := fmt.Sprintf("http://%s:%d/create/%s", serverConfig.Host, serverConfig.Port, *funcName)
	resp, err := postJson(url, requestBody)
	if err != nil {
		fmt.Printf("Creation request failed: %v", err)
		os.Exit(2)
	}
	printJsonResponse(resp.Body)
}

func postJson(url string, body []byte) (*http.Response, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("Server response: %v", resp.Status)
	}
	return resp, nil
}

func printJsonResponse(resp io.ReadCloser) {
	defer resp.Close()
	body, _ := ioutil.ReadAll(resp)

	// print indented JSON
	var out bytes.Buffer
	json.Indent(&out, body, "", "\t")
	out.WriteTo(os.Stdout)
}
