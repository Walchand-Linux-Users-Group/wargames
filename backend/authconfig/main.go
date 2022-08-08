package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	goHttp "net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/docker/docker/api/types/container"
	"go.containerssh.io/libcontainerssh/auth"
	authWebhook "go.containerssh.io/libcontainerssh/auth/webhook"
	"go.containerssh.io/libcontainerssh/config"
	configWebhook "go.containerssh.io/libcontainerssh/config/webhook"
	"go.containerssh.io/libcontainerssh/http"
	liblog "go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/metadata"
	"go.containerssh.io/libcontainerssh/service"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type authHandler struct {
}

func getImage(username string) string {

	postBody, _ := json.Marshal(map[string]string{
		"username": username,
		"apiToken": getEnv("API_TOKEN"),
	})

	responseBody := bytes.NewBuffer(postBody)

	resp, err := goHttp.Post(getEnv("API_URI")+"/status", "application/json", responseBody)

	if err != nil {
		log.Fatalln("An Error Occured %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type Response struct {
		Username     string `json:"username"`
		NextPassword string `json:"nextPassword"`
		Level        int    `json:"level"`
		Status       string `json:"status"`
	}

	var response Response
	json.Unmarshal(body, &response)

	fileName := strconv.Itoa(response.Level)
	nextPassword := response.NextPassword

	newName := "-" + fileName

	newName = response.Username + newName

	fmt.Println("FileName:")
	fmt.Println(string(body))

	content, err := ioutil.ReadFile(getEnv("IMAGE_FOLDER") + "/" + fileName)

	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string
	text := string(content)

	strings.ReplaceAll(text, "<{{{NEXT_PASSWORD}}}>", nextPassword)

	newFile, err := os.Create(getEnv("TMP_FOLDER") + "/" + newName)

	_, err2 := newFile.WriteString(text)

	if err2 != nil {
		log.Fatal(err2)
	}

	genImage(username, newName, getEnv("TMP_FOLDER")+"/"+newName)

	return newName
}

func genImage(username string, fileName string, filePath string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerFile := fileName
	dockerFileReader, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err, " :unable to open Dockerfile")
	}
	readDockerFile, err := ioutil.ReadAll(dockerFileReader)
	if err != nil {
		log.Fatal(err, " :unable to read dockerfile")
	}

	tarHeader := &tar.Header{
		Name: dockerFile,
		Size: int64(len(readDockerFile)),
	}
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, " :unable to write tar header")
	}
	_, err = tw.Write(readDockerFile)
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
	}
	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	imageBuildResponse, err := cli.ImageBuild(
		ctx,
		dockerFileTarReader,
		types.ImageBuildOptions{
			Context:    dockerFileTarReader,
			Dockerfile: dockerFile,
			Remove:     true,
			Tags:       []string{fileName + ":1.0"}})

	if err != nil {
		log.Fatal(err, " :unable to build docker image")
	}
	defer imageBuildResponse.Body.Close()
	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		log.Fatal(err, " :unable to read image build response")
	}
}

func verifyPassword(username string, password []byte) bool {
	pass := string(password[:])

	postBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": pass,
		"apiToken": getEnv("API_TOKEN"),
	})

	fmt.Println(string(postBody))

	responseBody := bytes.NewBuffer(postBody)

	resp, err := goHttp.Post(getEnv("API_URI")+"/auth", "application/json", responseBody)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type Response struct {
		Username string `json:"username"`
		Won      bool   `json:"won"`
		Level    int64  `json:"level"`
		Status   string `json:"status"`
	}

	var response Response
	fmt.Println(string(body))
	json.Unmarshal(body, &response)

	fmt.Println(response)

	return response.Status == "success"
}

func (a *authHandler) OnAuthorization(meta metadata.ConnectionAuthenticatedMetadata) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return true, meta.Authenticated(meta.Username), nil
}

func (a *authHandler) OnPubKey(meta metadata.ConnectionAuthPendingMetadata, publicKey auth.PublicKey) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return false, meta.AuthFailed(), nil
}

func (a *authHandler) OnPassword(metadata metadata.ConnectionAuthPendingMetadata, password []byte) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	if verifyPassword(metadata.Username, password) {
		return true, metadata.Authenticated(metadata.Username), nil
	}
	return false, metadata.AuthFailed(), nil
}

type configHandler struct {
}

func (c *configHandler) OnConfig(request config.Request) (config.AppConfig, error) {
	cfg := config.AppConfig{}

	cfg.Docker.Execution.Launch.ContainerConfig = &container.Config{}
	cfg.Docker.Execution.Launch.ContainerConfig.Image = getImage(request.Username)
	cfg.Docker.Execution.ImagePullPolicy = "Never"
	cfg.Docker.Execution.DisableAgent = true
	cfg.Docker.Execution.Mode = config.DockerExecutionModeSession
	cfg.Docker.Execution.ShellCommand = []string{"/bin/sh"}

	return cfg, nil
}

type handler struct {
	auth   goHttp.Handler
	config goHttp.Handler
}

func (h *handler) ServeHTTP(writer goHttp.ResponseWriter, request *goHttp.Request) {
	switch request.URL.Path {
	case "/password":
		fallthrough
	case "/pubkey":
		h.auth.ServeHTTP(writer, request)
	case "/config":
		h.config.ServeHTTP(writer, request)
	default:
		writer.WriteHeader(404)
	}
}

func initEnv() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func main() {
	initEnv()

	logger, err := liblog.NewLogger(
		config.LogConfig{
			Level:       config.LogLevelDebug,
			Format:      config.LogFormatLJSON,
			Destination: config.LogDestinationStdout,
		},
	)
	if err != nil {
		panic(err)
	}
	authHTTPHandler := authWebhook.NewHandler(&authHandler{}, logger)
	configHTTPHandler, err := configWebhook.NewHandler(&configHandler{}, logger)
	if err != nil {
		panic(err)
	}

	srv, err := http.NewServer(
		"authconfig",
		config.HTTPServerConfiguration{
			Listen: "0.0.0.0:8080",
		},
		&handler{
			auth:   authHTTPHandler,
			config: configHTTPHandler,
		},
		logger,
		func(s string) {

		},
	)
	if err != nil {
		panic(err)
	}

	running := make(chan struct{})
	stopped := make(chan struct{})
	lifecycle := service.NewLifecycle(srv)
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			println("Auth-Config Server is now running...")
			close(running)
		},
	).OnStopped(
		func(s service.Service, l service.Lifecycle) {
			close(stopped)
		},
	)
	exitSignalList := []os.Signal{os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM}
	exitSignals := make(chan os.Signal, 1)
	signal.Notify(exitSignals, exitSignalList...)
	go func() {
		if err := lifecycle.Run(); err != nil {
			panic(err)
		}
	}()
	select {
	case <-running:
		if _, ok := <-exitSignals; ok {
			println("Stopping Auth-Config Server...")
			lifecycle.Stop(context.Background())
		}
	case <-stopped:
	}
}
