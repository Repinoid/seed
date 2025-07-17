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

	req := testcontainers.ContainerRequest{
		Image:        "naeel/iman:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp").WithStartupTimeout(120 * time.Second),
	}
	var err error
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
