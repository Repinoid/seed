package tests

import (
	"context"
	"database/sql"
	"fmt"
	"gomuncool/internal/models"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq" // <- Важно: этот импорт регистрирует драйвер
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

// выполняется перед тестами
func (suite *TstSeed) SetupSuite() {
	suite.ctx = context.Background()
	suite.t = time.Now()

	os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")

	var err error

	//  1. Create network (simplified modern API)
	suite.testNet, err = network.New(suite.ctx,
		network.WithAttachable(),
		network.WithLabels(map[string]string{
			"test": "handlers-suite",
		}),
	)
	if err != nil {
		suite.FailNowf("Failed to create network: %v", err.Error())
	}

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
		Networks: []string{suite.testNet.Name},
	}

	postgresContainer, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)

	pgIP, err := postgresContainer.ContainerIP(suite.ctx)
	_ = pgIP
	suite.Require().NoError(err)

	// Получение хоста и порта postgres
	suite.pgHost, err = postgresContainer.Host(suite.ctx)
	suite.Require().NoError(err)
	// get externally mapped port for a container port
	suite.pgPort, err = postgresContainer.MappedPort(suite.ctx, "5432")
	suite.Require().NoError(err)
	//suite.DBEndPoint = fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", suite.pgHost, "5432")
	suite.DBEndPoint = fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", suite.pgHost, suite.pgPort.Port())
	models.DBEndPoint = suite.DBEndPoint
	suite.postgresContainer = postgresContainer
	models.Logger.Info("PostgreSQL доступен по адресу: ",
		"Host", suite.pgHost,
		"Port", suite.pgPort.Port())

	// Дополнительная проверка
	db, err := sql.Open("postgres", models.DBEndPoint)
	suite.Require().NoError(err)
	db.Close()

	spr := fmt.Sprintf("host=%s port=%d user=testuser password=testpass dbname=testdb sslmode=disable", suite.pgHost, suite.pgPort.Int())
	//spr := fmt.Sprintf("host=%s port=%d user=testuser password=testpass dbname=testdb sslmode=disable", "go_db", suite.pgPort.Int())
	db, err = sql.Open("postgres", spr)
	suite.Require().NoError(err)
	db.Close()

	models.DBEndPoint = spr
	models.Logger.Info("PostGres GenericContainer Spent ", "", time.Since(suite.t))

	// ***************** POSTGREs part end ************************************

	models.Logger.Info("PostGres ", "EndPoint", models.DBEndPoint)

	// ***************** IMANs part begin ************************************

	time.Sleep(10*time.Second)

	suite.servakContainer, err = testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			// FromDockerfile: testcontainers.FromDockerfile{
			// 	Context:    "../../../",
			// 	Dockerfile: "ServerDockerFile",
			// },
			Image: "iman:1",
			//Image:        "naeel/iman:latest",
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				//"DB_HOST":      suite.pgHost,
				//"DB_PORT":      suite.pgPort.Port(),
				// "DB_HOST":      pgIP,
				// "DB_PORT":      "5432",
				// "DB_USER":      "testusera",
				// "DB_PASSWORD":  "testpassa",
				// "DB_NAME":      "testdba",
				"DATABASE_DSN": models.DBEndPoint,
			},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("8080/tcp").WithStartupTimeout(30*time.Second),
				wait.ForHTTP("/health").WithPort("8080/tcp").WithStartupTimeout(30*time.Second),
				wait.ForLog("HTTPa server started"),
			).WithDeadline(10 * time.Second), //
			Networks: []string{suite.testNet.Name},
			// HostConfigModifier: func(hostConfig *container.HostConfig) {
			// 	hostConfig.PortBindings = nat.PortMap{
			// 		"8080/tcp": []nat.PortBinding{
			// 			{
			// 				HostIP:   "0.0.0.0",
			// 				HostPort: "8080",
			// 			},
			// 		},
			// 	}
			// },
		},
		Started: true,
		Reuse:   false,
	})
	suite.Assert().NoError(err)

	models.Logger.Info("Iman's Spent ", "", time.Since(suite.t))

	logsBytes, err := suite.servakContainer.Logs(context.Background())
	if err != nil {
		suite.T().Fatal("Failed to get container logs:", err)
	}
	defer logsBytes.Close() // Important!

	// Convert to string
	logs, err := io.ReadAll(logsBytes)
	if err != nil {
		suite.T().Fatal("Failed to read container logs:", err)
	}

	// Print or assert
	fmt.Println("Container logs:", string(logs))
	// Or in a test failure:
	suite.T().Log("Container logs:", string(logs))

	fmt.Println(string(logs))
	suite.Require().NoError(err)

	// ***************** IMANs part end ************************************

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
