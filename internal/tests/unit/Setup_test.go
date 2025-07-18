package tests

import (
	"context"
	"fmt"
	"gomuncool/internal/models"
	"io"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq" // <- Важно: этот импорт регистрирует драйвер
)

//var hostName = "localhost"

var hostName = "dbhost"

// выполняется перед тестами
func (suite *TstSeed) SetupSuite() {
	suite.ctx = context.Background()
	suite.t = time.Now()

	//os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock") на эту херь ругается github actions

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

	// прописываем postgresContainer в переменные TstSeed struct чтобы по итогу suite.postgresContainer.Terminate(suite.ctx)
	suite.postgresContainer = postgresContainer

	// Получение хоста и порта postgres
	suite.pgHost, err = postgresContainer.Host(suite.ctx)
	suite.Require().NoError(err)
	// get externally mapped port for a container port
	// Because the randomised port mapping happens during container startup, the container must be running at the time MappedPort is called.
	// You may need to ensure that the startup order of components in your tests caters for this.
	suite.pgPort, err = postgresContainer.MappedPort(suite.ctx, "5432")
	suite.Require().NoError(err)

	// suite.DBEndPoint используется тестами.
	// его хост/порт определяется через postgresContainer.Host и postgresContainer.MappedPort(suite.ctx, "5432")
	suite.DBEndPoint = fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb", suite.pgHost, suite.pgPort.Port())

	models.Logger.Info("PostgreSQL доступен по адресу: ",
		"Host", suite.pgHost,
		"Port", suite.pgPort.Port())

	// в этом эндпоинте хост - имя контейнера постгреса, порт 5432. Используется для контейнера приложения, по нему он в базу стучится
	models.DBEndPoint = "host=" + hostName + " port=" + "5432" + " user=testuser password=testpass dbname=testdb sslmode=disable"

	models.Logger.Info("PostGres GenericContainer Spent ", "", time.Since(suite.t))

	// ***************** POSTGREs part end ************************************

	models.Logger.Info("PostGres ", "EndPoint", models.DBEndPoint)

	// ***************** IMANs part begin ************************************

	requ := testcontainers.ContainerRequest{
		//Image: "iman:1",
		Image:        "naeel/iman:latest",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"DATABASE_DSN": models.DBEndPoint,
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("8080/tcp").WithStartupTimeout(60*time.Second),
			wait.ForLog("HTTP server started"),
		).WithDeadline(90 * time.Second), //
		Networks: []string{suite.testNet.Name},
	}

	suite.servakContainer, err = testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: requ,
		Started:          true,
		Reuse:            false,
	})
	suite.Assert().NoError(err)

	models.Logger.Info("Iman's Spent ", "", time.Since(suite.t))

	// Получение хоста и порта сервера
	suite.servakHost, err = suite.servakContainer.Host(suite.ctx)
	suite.Require().NoError(err)
	// get externally mapped port for a container port
	// Because the randomised port mapping happens during container startup, the container must be running at the time MappedPort is called.
	// You may need to ensure that the startup order of components in your tests caters for this.
	suite.servakPort, err = suite.servakContainer.MappedPort(suite.ctx, "8080")
	suite.Require().NoError(err)

	// вывод логов контейнера, отлаживал
	logsBytes, err := suite.servakContainer.Logs(context.Background())
	if err != nil {
		suite.T().Fatal("Failed to get container logs:", err)
	}
	defer logsBytes.Close()

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
