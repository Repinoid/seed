package tests

import (
	"context"
	"gomuncool/internal/models"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TstSeed struct {
	suite.Suite
	t               time.Time
	ctx             context.Context
	servakContainer testcontainers.Container
	host            string
	port            nat.Port
}

// выполняется перед тестами
func (suite *TstSeed) SetupSuite() {
	suite.ctx = context.Background()
	suite.t = time.Now()

	os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")

	// 1. Create network (simplified modern API)
	testNet, err := network.New(suite.ctx,
		network.WithAttachable(),
		network.WithLabels(map[string]string{
			"test": "handlers-suite",
		}),
	)
	if err != nil {
		suite.FailNowf("Failed to create network: %v", err.Error())
	}
	defer func() {
		if err := testNet.Remove(suite.ctx); err != nil {
			suite.T().Logf("Network cleanup warning: %v", err)
		}
	}()

	// 2. Start PostgreSQL with modern options
	pgContainer, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{ // Preferred over Env in newer versions
				"POSTGRES_USER":     "uname",
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       "dbase",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(2 * time.Minute).
				WithPollInterval(1 * time.Second),
			Networks: []string{testNet.Name},
		},
		Started: true,
	})
	if err != nil {
		suite.FailNowf("Failed to start PostgreSQL: %v", err.Error())
	}
	defer func() {
		if err := pgContainer.Terminate(suite.ctx); err != nil {
			suite.T().Logf("PostgreSQL cleanup warning: %v", err)
		}
	}()

	// 3. Start application container
	appContainer, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "naeel/iman:latest",
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				"DB_HOST":     "postgres",
				"DB_PORT":     "5432",
				"DB_USER":     "uname",
				"DB_PASSWORD": "password",
				"DB_NAME":     "dbase",
			},
			WaitingFor: wait.ForHTTP("/health").
				WithPort("8080/tcp").
				WithStartupTimeout(2 * time.Minute).
				WithPollInterval(1 * time.Second),
			Networks: []string{testNet.Name},
			// LogConsumerCfg: &testcontainers.LogConsumerConfig{
			// 	Opts: []testcontainers.LogConsumerOption{
			// 		testcontainers.WithStdoutLogs(),
			// 		testcontainers.WithStderrLogs(),
			// 	},
			//},
		},
		Started: true,
	})
	if err != nil {
		suite.FailNowf("Failed to start application: %v", err.Error())
	}
	defer func() {
		if err := appContainer.Terminate(suite.ctx); err != nil {
			suite.T().Logf("App container cleanup warning: %v", err)
		}
	}()

	req := testcontainers.ContainerRequest{
		Image:        "naeel/iman:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp").WithStartupTimeout(120 * time.Second),
	}
	suite.servakContainer, err = testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)
	// Получение хоста и порта
	suite.host, err = suite.servakContainer.Host(suite.ctx)
	suite.Require().NoError(err)
	// var
	suite.port, err = suite.servakContainer.MappedPort(suite.ctx, "8080")
	suite.Require().NoError(err)

}

func (suite *TstSeed) TearDownSuite() { // // выполняется после всех тестов
	models.Logger.Info("Spent ", "", time.Since(suite.t))
	suite.servakContainer.Terminate(suite.ctx)
}

func TestHandlersSuite(t *testing.T) {
	testBase := new(TstSeed)
	testBase.ctx = context.Background()

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Минимальный уровень логирования
		AddSource: true,            // Добавлять информацию об исходном коде

	})
	models.Logger = slog.New(handler)
	slog.SetDefault(models.Logger)

	models.Logger.Info("before run ....")
	suite.Run(t, testBase)

}
