package tests

import (
	"context"
	"fmt"
	"gomuncool/internal/models"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TstSeed struct {
	suite.Suite
	t   time.Time
	ctx context.Context
	//	servakContainer testcontainers.Container

	//host string
	//port nat.Port

	DBEndPoint        string
	postgresContainer testcontainers.Container
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

	// ***************** POSTGREs part begin ************************************
	// Запуск контейнера PostgreSQL
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(2 * time.Minute).
			WithPollInterval(1 * time.Second),
		Networks: []string{testNet.Name},
	}

	postgresContainer, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)
	//	defer postgresContainer.Terminate(suite.ctx)

	// Получение хоста и порта
	host, err := postgresContainer.Host(suite.ctx)
	suite.Require().NoError(err)
	port, err := postgresContainer.MappedPort(suite.ctx, "5432")
	suite.Require().NoError(err)
	suite.DBEndPoint = fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", host, port.Port())
	suite.postgresContainer = postgresContainer
	models.Logger.Info("PostgreSQL доступен по адресу: %s:%s", host, port.Port())

	// ***************** POSTGREs part end ************************************

}

// TearDownSuite выполняется после всех тестов
func (suite *TstSeed) TearDownSuite() {
	// Вывод времени исполнения тестов
	models.Logger.Info("Spent ", "", time.Since(suite.t))
	// убиваем контейнер постгреса
	suite.postgresContainer.Terminate(suite.ctx)
}

func TestHandlersSuite(t *testing.T) {
	testBase := new(TstSeed)
	testBase.ctx = context.Background()

	// вывод в os.Stdout
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Минимальный уровень логирования
		AddSource: true,            // Добавлять информацию об исходном коде

	})
	models.Logger = slog.New(handler)
	slog.SetDefault(models.Logger)

	models.Logger.Info("before run ....")
	suite.Run(t, testBase)

}
