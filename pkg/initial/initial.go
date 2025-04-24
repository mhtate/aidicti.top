package initial

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"math/big"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"aidicti.top/pkg/discovery"
	"aidicti.top/pkg/discovery/consul"
	"aidicti.top/pkg/logging"
	logslog "aidicti.top/pkg/logging/slog"
	"aidicti.top/pkg/utils"
	"github.com/go-git/go-git/v5"
)

func getStdoutLogHandler(opt *slog.HandlerOptions) *slog.JSONHandler {
	return slog.NewJSONHandler(os.Stdout, opt)
}

func getFileLogHandler(fileName string, opt *slog.HandlerOptions) (*slog.TextHandler, func() error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	utils.Assert(err == nil, fmt.Sprintf("open log file fail, \"name\":\"%s\"", fileName))

	return slog.NewTextHandler(file, opt), func() error { return file.Close() }
}

// getInstanceID generates a pseudo-unique service instance identifier, using a service name
// suffixed by dash and a random number.
func getInstanceID(serviceName string) string {
	const IDruneSize = 3
	const HashruneSize = 3

	generateBase64ID := func() string {
		uintId := rand.New(rand.NewSource(time.Now().UnixNano())).Uint64()

		// Convert uint64 to bytes
		buf := new(big.Int).SetUint64(uintId).Bytes()

		for {
			// Encode to base64
			encoded := base64.RawURLEncoding.EncodeToString(buf)

			if len(encoded) > IDruneSize {
				return encoded[:IDruneSize]
			}
		}
	}

	getGitHashFromSource := func() (string, error) {
		ex, err := os.Executable()
		utils.Assert(err == nil, "get running file fail")

		getGitRoot := func(path string) (string, error) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return "", err
			}

			for {
				_, err := git.PlainOpen(absPath)
				if err == nil {
					return absPath, nil // Found the root of the Git repository
				}

				// Move up one directory
				parent := filepath.Dir(absPath)
				if parent == absPath { // If we reach the filesystem root, stop
					return "", fmt.Errorf("no Git repository found")
				}
				absPath = parent
			}
		}

		rootPath, err := getGitRoot(ex)
		if err != nil {
			logging.Debug("get git hash fail", "err", err)
			return "", err
		}

		repo, err := git.PlainOpen(rootPath)
		if err != nil {
			logging.Debug("get git hash fail", "err", err)
			return "", err
		}

		ref, err := repo.Head()
		if err != nil {
			logging.Debug("get git hash fail", "err", err)
			return "", err
		}

		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			logging.Debug("get git hash fail", "err", err)
			return "", err
		}

		logging.Info("get git hash ok", "hash", commit.Hash)

		return commit.Hash.String(), nil
	}

	getGitHashFromEnv := func() (string, error) {
		const CommitHashPathEnv = "AIDICTI_COMMIT_HASH"
		value := os.Getenv(CommitHashPathEnv)
		if value == "" {
			return "", fmt.Errorf("get env var fail", "name", CommitHashPathEnv)
		}

		return value, nil
	}

	getGitHashFromFile := func() (string, error) {
		data, err := ioutil.ReadFile("./commit_hash")
		if err != nil {
			return "", fmt.Errorf("get env var fail", "name", "commit_hash")
		}

		return string(data), nil
	}

	hash, err := getGitHashFromSource()
	if err == nil {
		return fmt.Sprintf("%s-%s%s", serviceName, generateBase64ID(), hash[:HashruneSize])
	}

	hash, err = getGitHashFromEnv()
	if err == nil {
		return fmt.Sprintf("%s-%s%s", serviceName, generateBase64ID(), hash[:HashruneSize])
	}

	hash, err = getGitHashFromFile()
	if err == nil {
		return fmt.Sprintf("%s-%s%s", serviceName, generateBase64ID(), hash[:HashruneSize])
	}

	utils.Assert(false, "get hash fail")
	return ""
}

func getMyIP(sameSubnetAddr string) (string, error) {
	conn, err := net.Dial("udp", sameSubnetAddr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func Init(serviceName string, port uint64, init func(discovery.Registry, string)) {
	registryName := flag.String("reg", "localhost:8500", "Registry domain name")
	isLevelDebug := flag.Bool("debug", false, "Enable debug log level")
	logFile := flag.String("logfile", "", "Additional LogFile (Stdout still enabled)")
	flag.Parse()

	options := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// ReplaceAttr: logslog.AddSourceIndent(func(s []string, a slog.Attr) slog.Attr { return a }),
	}

	if *isLevelDebug {
		options.Level = slog.LevelDebug
	}

	defaultHandler := getStdoutLogHandler(options)
	logging.Set(slog.New(defaultHandler))

	instanceId := getInstanceID(serviceName)

	addrLocal, err := getMyIP(*registryName)
	utils.Assert(err == nil, "get local addr fail")

	addrPort := fmt.Sprintf("%s:%d", addrLocal, port)

	serviceData := slog.GroupValue(
		slog.String("name", serviceName),
		slog.String("id", instanceId),
		slog.String("addr", addrPort))

	logging.Set(slog.New(defaultHandler).With("service", serviceData))

	logging.Info("get flag data",
		"registryName", *registryName,
		"isLevelDebug", isLevelDebug,
		"logFile", *logFile,
	)

	registry, err := consul.NewRegistry(*registryName)
	utils.Assert(err == nil, "find consul registry failed")

	deregister, instanceId, err := discovery.RegisterService(
		context.Background(), serviceName, instanceId, addrPort, registry)

	utils.Assert(err == nil, "register consul registry failed")
	defer deregister()

	if *logFile != "" {
		fileHandler, close := logslog.NewLocalFileHandler(instanceId, options)

		logging.Set(
			slog.New(
				logslog.NewMultiHandler(fileHandler, defaultHandler)).With("service", serviceData))

		defer close()
	}

	init(registry, addrPort)
}
