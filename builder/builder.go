package builder

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alexellis/hmac/v2"
)

const BuilderConfigFileName = "com.openfaas.docker.config"

type BuildConfig struct {
	// Image reference.
	Image string `json:"image"`

	// Extra build arguments for the Dockerfile.
	BuildArgs map[string]string `json:"buildArgs,omitempty"`

	// Platforms for multi-arch builds.
	Platforms []string `json:"platforms,omitempty"`
}

type BuildResult struct {
	Log    []string `json:"log"`
	Image  string   `json:"image"`
	Status string   `json:"status"`
}

type FunctionBuilder struct {
	// URL of the OpenFaaS Builder API.
	URL *url.URL

	// Http client used for calls to the builder API.
	client *http.Client

	// HMAC secret used for hashing request payloads.
	hmacSecret string
}

type BuilderOption func(*FunctionBuilder)

// WithHmacAuth configures the HMAC secret used to sign request payloads to the builder API.
func WithHmacAuth(secret string) BuilderOption {
	return func(b *FunctionBuilder) {
		b.hmacSecret = secret
	}
}

// NewFunctionBuilder create a new builder for building OpenFaaS functions using the Function Builder API.
func NewFunctionBuilder(url *url.URL, client *http.Client, options ...BuilderOption) *FunctionBuilder {
	b := &FunctionBuilder{
		URL: url,

		client: client,
	}

	for _, option := range options {
		option(b)
	}

	return b
}

// Build invokes the function builder API with the provided tar archive containing the build config and context
// to build and push a function image.
func (b *FunctionBuilder) Build(tarPath string) (BuildResult, error) {
	tarFile, err := os.Open(tarPath)
	if err != nil {
		return BuildResult{}, err
	}
	defer tarFile.Close()

	tarFileBytes, err := io.ReadAll(tarFile)
	if err != nil {
		return BuildResult{}, err
	}

	u := b.URL.JoinPath("/build")

	digest := hmac.Sign(tarFileBytes, []byte(b.hmacSecret), sha256.New)
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(tarFileBytes))
	if err != nil {
		return BuildResult{}, err
	}

	req.Header.Set("X-Build-Signature", "sha256="+hex.EncodeToString(digest))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", "openfaas-go-sdk")

	res, err := b.client.Do(req)
	if err != nil {
		return BuildResult{}, err
	}

	result := BuildResult{}
	if res.Body != nil {
		defer res.Body.Close()

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return BuildResult{}, err
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return BuildResult{}, err
		}
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return result, fmt.Errorf("failed to build function, builder responded with status code %d, build status: %s", res.StatusCode, result.Status)
	}

	return result, nil
}

// MakeTar create a tar archive that contains the build config and build context.
func MakeTar(tarPath string, context string, buildConfig *BuildConfig) error {
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return err
	}

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	if err = filepath.Walk(context, func(path string, f os.FileInfo, pathErr error) error {
		if pathErr != nil {
			return pathErr
		}

		targetFile, err := os.Open(path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(f, f.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.Join("context", strings.TrimPrefix(path, context))
		header.Name = strings.TrimPrefix(header.Name, "/")

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if f.Mode().IsDir() {
			return nil
		}

		_, err = io.Copy(tarWriter, targetFile)
		return err
	}); err != nil {
		return err
	}

	configBytes, err := json.Marshal(buildConfig)
	if err != nil {
		return err
	}

	configHeader := &tar.Header{
		Name: BuilderConfigFileName,
		Mode: 0664,
		Size: int64(len(configBytes)),
	}

	if err := tarWriter.WriteHeader(configHeader); err != nil {
		return err
	}

	if _, err := tarWriter.Write(configBytes); err != nil {
		return err
	}

	return err
}

const (
	DefaultTemplateDir     = "./template"
	DefaultTemplateHandler = "function"
	DefaultBuildDir        = "./build"
)

type BuildContextOption func(*BuildContextConfig)

type BuildContextConfig struct {
	// Directory where the build context will be created.
	BuildDir string

	// Directory used to lookup templates
	TemplateDir string

	// Path where the function handler should be overlayed
	// in the selected template
	TemplateHandlerOverlay string
}

// WithBuildDir is an option to configure the directory the build context is created in.
// If this options is not set a default path `./build` is used.
func WithBuildDir(path string) BuildContextOption {
	return func(c *BuildContextConfig) {
		c.BuildDir = path
	}
}

// WithTemplateDir is an option to configure the directory where the build
// template is looked up.
// If this option is not set a default path `./template` is used.
func WithTemplateDir(path string) BuildContextOption {
	return func(c *BuildContextConfig) {
		c.TemplateDir = path
	}
}

// WithHandlerOverlay is an option to configure the path where the function handler needs to be
// overlayed in the template.
// If this option is not set a default overlay path `function` is used.
func WithHandlerOverlay(path string) BuildContextOption {
	return func(c *BuildContextConfig) {
		c.TemplateHandlerOverlay = path
	}
}

// CreateBuildContext create a Docker build context using the provided function handler and language template.
//
// Parameters:
//   - functionName: name of the function.
//   - handler: path to the function handler.
//   - language: name of the language template to use.
//   - copyExtraPaths: additional paths to copy into the function handler folder. Paths should be relative to the current directory.
//     Any paths outside if this directory will be skipped.
//
// By default templates are looked up in the `./template` directory. The path the the template
// directory can be overridden by setting the `builder.WithTemplateDir` option.
// CreateBuildContext overlays the function handler in the `function` folder of the template by default.
// This setting can be overridden by setting the `builder.WithHandlerOverlay` option.
//
// The function returns the path to the build context, `./build/<functionName>` by default.
// The build directory can be overridden by setting the `builder.WithBuildDir` option.
// An error is returned if creating the build context fails.
func CreateBuildContext(functionName string, handler string, language string, copyExtraPaths []string, options ...BuildContextOption) (string, error) {
	c := &BuildContextConfig{
		BuildDir:               DefaultBuildDir,
		TemplateHandlerOverlay: DefaultTemplateHandler,
		TemplateDir:            DefaultTemplateDir,
	}

	for _, option := range options {
		option(c)
	}

	contextPath := path.Join(c.BuildDir, functionName)

	if err := os.RemoveAll(contextPath); err != nil {
		return contextPath, fmt.Errorf("unable to clear context folder: %s", contextPath)
	}

	handlerDst := contextPath
	if language != "dockerfile" {
		handlerDst = path.Join(contextPath, c.TemplateHandlerOverlay)
	}

	permissions := defaultDirPermissions
	if isRunningInCI() {
		permissions = 0777
	}

	err := os.MkdirAll(handlerDst, permissions)
	if err != nil {
		return contextPath, fmt.Errorf("error creating function handler path %s: %w", handlerDst, err)
	}

	if language != "dockerfile" {
		templateSrc := path.Join(c.TemplateDir, language)
		if err := copyFiles(templateSrc, contextPath); err != nil {
			return contextPath, fmt.Errorf("error copying template %s: %w", language, err)
		}
	}

	// Overlay function handler in template.
	handlerSrc := handler
	infos, err := os.ReadDir(handlerSrc)
	if err != nil {
		return contextPath, fmt.Errorf("error reading function handler %s: %w", handlerSrc, err)
	}

	for _, info := range infos {
		switch info.Name() {
		case "build", "template":
			continue
		default:
			if err := copyFiles(
				filepath.Clean(path.Join(handlerSrc, info.Name())),
				filepath.Clean(path.Join(handlerDst, info.Name())),
			); err != nil {
				return contextPath, err
			}
		}
	}

	for _, extraPath := range copyExtraPaths {
		extraPathAbs, err := pathInScope(extraPath, ".")
		if err != nil {
			return contextPath, err
		}
		// Note that if template is nil or the language is `dockerfile`, then
		// handlerDest == contextPath, the docker build context, not the handler folder
		// inside the docker build context.
		if err := copyFiles(
			extraPathAbs,
			filepath.Clean(path.Join(handlerDst, extraPath)),
		); err != nil {
			return contextPath, fmt.Errorf("error copying extra paths: %w", err)
		}
	}

	return contextPath, nil
}

// pathInScope returns the absolute path to `path` and ensures that it is located within the
// provided scope. An error will be returned, if the path is outside of the provided scope.
func pathInScope(path string, scope string) (string, error) {
	scope, err := filepath.Abs(filepath.FromSlash(scope))
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(filepath.FromSlash(path))
	if err != nil {
		return "", err
	}

	if abs == scope {
		return "", fmt.Errorf("forbidden path appears to equal the entire project: %s (%s)", path, abs)
	}

	if strings.HasPrefix(abs, scope) {
		return abs, nil
	}

	// default return is an error
	return "", fmt.Errorf("forbidden path appears to be outside of the build context: %s (%s)", path, abs)
}

const defaultDirPermissions os.FileMode = 0700

// isRunningInCI checks the ENV var CI and returns true if it's set to true or 1
func isRunningInCI() bool {
	if env, ok := os.LookupEnv("CI"); ok {
		if env == "true" || env == "1" {
			return true
		}
	}
	return false
}
