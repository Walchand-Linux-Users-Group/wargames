/*
main package is the main entry point for the authconfig backend.
*/
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	goHttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Walchand-Linux-Users-Group/wargames/backend/authconfig/helpers"
	"github.com/docker/docker/api/types/container"
	"go.containerssh.io/libcontainerssh/auth"
	authWebhook "go.containerssh.io/libcontainerssh/auth/webhook"
	"go.containerssh.io/libcontainerssh/config"
	configWebhook "go.containerssh.io/libcontainerssh/config/webhook"
	"go.containerssh.io/libcontainerssh/http"
	liblog "go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/metadata"
	"go.containerssh.io/libcontainerssh/service"
)

type authHandler struct {
}

func getImage(username string) string {

	postBody, _ := json.Marshal(map[string]string{
		"username": username,
		"apiToken": helpers.GetEnv("API_TOKEN"),
	})

	responseBody := bytes.NewBuffer(postBody)

	resp, err := goHttp.Post(helpers.GetEnv("API_URI")+"/image", "application/json", responseBody)

	if err != nil {
		fmt.Println("Wargames API seems down!")
		return ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error in Verifying User or API compatibility issue!")
		return ""
	}

	type Image struct {
		Level            int64  `json:"level"`
		ImageName        string `json:"imageName"`
		ImageRegistryURL string `json:"imageRegistryURL"`
		ImageDesc        string `json:"imageDesc"`
		Flag             string `json:"flag"`
		Status           string `json:"status"`
	}

	var img Image
	json.Unmarshal(body, &img)

	return img.ImageRegistryURL
}

func verifyUser(username string) bool {

	postBody, _ := json.Marshal(map[string]string{
		"username": username,
	})

	fmt.Println(string(postBody))

	responseBody := bytes.NewBuffer(postBody)

	resp, err := goHttp.Post(helpers.GetEnv("API_URI")+"/stats", "application/json", responseBody)

	if err != nil {
		fmt.Println("Wargames API seems down! - Verifying User")
		return false
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error in Verifying User or API compatibility issue!")
		return false
	}

	type Stat struct {
		Timestamp int64 `json:"timestamp"`
		Level     int64 `json:"level"`
	}

	type User struct {
		Username  string `json:"username"`
		Name      string `json:"name"`
		Level     int64  `json:"level"`
		Org       string `json:"org"`
		Timestamp int64  `json:"timestamp"`
		Stats     []Stat `json:"stats"`
		Status    string `json:"status"`
	}

	var response User
	json.Unmarshal(body, &response)

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
	if verifyUser(metadata.Username) {
		fmt.Println("SSH successful for username ", metadata.Username)
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
	cfg.Docker.Execution.ImagePullPolicy = "IfNotPresent"
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

/*
AuthConfig handles intermediate SSH authentication.
*/
func main() {
	helpers.InitEnv()

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
