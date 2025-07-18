package tests

import (
	"context"
	"database/sql"
	"fmt"
	"gomuncool/internal/models"
	"io"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq" // <- Важно: этот импорт регистрирует драйвер
)

var hostName = "dbhost"

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
	suite.Require().NoError(err)

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
		NetworkAliases: map[string][]string{
			suite.testNet.Name: {hostName}, // <-- Explicit alias
		},
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

	//spr := fmt.Sprintf("host=%s port=%d user=testuser password=testpass dbname=testdb sslmode=disable", suite.pgHost, suite.pgPort.Int())
	spr := fmt.Sprintf("host=%s port=%d user=testuser password=testpass dbname=testdb sslmode=disable", hostName, suite.pgPort.Int())
	db, err = sql.Open("postgres", spr)
	suite.Require().NoError(err)
	db.Close()

	models.DBEndPoint = spr
	models.Logger.Info("PostGres GenericContainer Spent ", "", time.Since(suite.t))

	// ***************** POSTGREs part end ************************************

	models.Logger.Info("PostGres ", "EndPoint", models.DBEndPoint)

	// ***************** IMANs part begin ************************************

	//time.Sleep(10 * time.Second)

	suite.servakContainer, err = testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "iman:1",
			//Image:        "naeel/iman:latest",
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				//"DATABASE_DSN": models.DBEndPoint,
				// Use "postgres" (container name) instead of "localhost"
				"DATABASE_DSN": "host=" + hostName + " port=" + suite.pgPort.Port() + " user=testuser password=testpass dbname=testdb sslmode=disable",
			},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("8080/tcp").WithStartupTimeout(60*time.Second),
				wait.ForHTTP("/health").WithPort("8080/tcp").WithStartupTimeout(60*time.Second),
				wait.ForLog("HTTP server started"),
			).WithDeadline(90 * time.Second), //
			Networks: []string{suite.testNet.Name},
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
