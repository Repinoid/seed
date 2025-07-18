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
	//_ "github.com/lib/pq" // <- Важно: этот импорт регистрирует драйвер
)

type TstSeed struct {
	suite.Suite
	t   time.Time
	ctx context.Context

	testNet *testcontainers.DockerNetwork

	servakContainer testcontainers.Container

	pgHost string
	pgPort nat.Port

	DBEndPoint        string
	postgresContainer testcontainers.Container
}

// TearDownSuite выполняется после всех тестов
func (suite *TstSeed) TearDownSuite() {
	// Вывод времени исполнения тестов
	models.Logger.Info("Spent ", "", time.Since(suite.t))

	// убиваем контейнер постгреса
	err := suite.postgresContainer.Terminate(suite.ctx)
	suite.Assert().NoError(err)

	// убиваем контейнер IMAN
	err = suite.servakContainer.Terminate(suite.ctx)
	suite.Assert().NoError(err)

	// kill network
	err = suite.testNet.Remove(suite.ctx)
	suite.Assert().NoError(err)

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
