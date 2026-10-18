//go:build e2e

package e2e

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// labrinthURL is the base URL of the labrinth API, without a trailing slash.
var labrinthURL string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	start := time.Now()
	s := &stack{}
	defer func() {
		if err := s.terminate(); err != nil {
			log.Printf("teardown: %v", err)
		}
	}()

	if err := s.start(context.Background()); err != nil {
		log.Printf("start stack: %v", err)
		return 1
	}
	log.Printf("stack ready after %s", time.Since(start).Round(100*time.Millisecond))

	return m.Run()
}

type stack struct {
	network *testcontainers.DockerNetwork
	// volume holds labrinth's uploaded files, served by the cdn container.
	volume     string
	containers []testcontainers.Container
}

func (s *stack) start(ctx context.Context) error {
	nw, err := network.New(ctx)
	if err != nil {
		return fmt.Errorf("create network: %w", err)
	}
	s.network = nw
	s.volume = nw.Name + "-cdn"

	_, err = s.runContainer(ctx, "postgres", imagePostgres,
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     "labrinth",
			"POSTGRES_PASSWORD": "labrinth",
		}),
		// The entrypoint restarts postgres once after initdb.
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	if err != nil {
		return err
	}

	_, err = s.runContainer(ctx, "redis", imageRedis,
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
	)
	if err != nil {
		return err
	}

	_, err = s.runContainer(ctx, "clickhouse", imageClickhouse,
		testcontainers.WithEnv(map[string]string{
			"CLICKHOUSE_DB":       "staging_ariadne",
			"CLICKHOUSE_USER":     "default",
			"CLICKHOUSE_PASSWORD": "default",
		}),
		testcontainers.WithExposedPorts("8123/tcp"),
		// The entrypoint creates the database on a server bound to
		// localhost, so the mapped port answers only once the real
		// server is up.
		testcontainers.WithWaitStrategy(wait.ForHTTP("/ping").WithPort("8123/tcp")),
	)
	if err != nil {
		return err
	}

	_, err = s.runContainer(ctx, "redpanda", imageRedpanda,
		testcontainers.WithCmd("redpanda", "start",
			"--mode", "dev-container", "--smp", "1",
			"--kafka-addr", "0.0.0.0:9092",
			"--advertise-kafka-addr", "redpanda:9092"),
		testcontainers.WithWaitStrategy(wait.ForExec([]string{"rpk", "cluster", "info", "-X", "brokers=localhost:9092"})),
	)
	if err != nil {
		return err
	}

	// labrinth's local storage backend writes files but never serves them.
	cdn, err := s.runContainer(ctx, "cdn", imageBusybox,
		testcontainers.WithCmd("httpd", "-f", "-p", "80", "-h", "/srv"),
		testcontainers.WithMounts(testcontainers.ContainerMount{
			Source:   testcontainers.GenericVolumeMountSource{Name: s.volume},
			Target:   "/srv",
			ReadOnly: true,
		}),
		testcontainers.WithExposedPorts("80/tcp"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("80/tcp")),
	)
	if err != nil {
		return err
	}
	cdnPort, err := cdn.MappedPort(ctx, "80/tcp")
	if err != nil {
		return fmt.Errorf("cdn port: %w", err)
	}

	env := map[string]string{
		"DATABASE_URL":            "postgresql://labrinth:labrinth@postgres/labrinth",
		"REDIS_URL":               "redis://redis",
		"CLICKHOUSE_URL":          "http://clickhouse:8123",
		"KAFKA_BOOTSTRAP_SERVERS": "redpanda:9092",
	}

	migrate, err := s.runContainer(ctx, "migrate", imageLabrinth,
		testcontainers.WithEnv(env),
		testcontainers.WithCmd("/labrinth/labrinth", "--run-background-task", "migrations"),
		testcontainers.WithWaitStrategy(wait.ForExit().WithExitTimeout(2*time.Minute)),
	)
	if err != nil {
		return err
	}
	state, err := migrate.State(ctx)
	if err != nil {
		return fmt.Errorf("migrate state: %w", err)
	}
	if state.ExitCode != 0 {
		return fmt.Errorf("migrate exited with %d:\n%s", state.ExitCode, tail(ctx, migrate))
	}

	labrinth, err := s.runContainer(ctx, "labrinth", imageLabrinth,
		testcontainers.WithEnv(env),
		testcontainers.WithEnv(map[string]string{
			"BIND_ADDR":                   "0.0.0.0:8000",
			"CDN_URL":                     "http://localhost:" + cdnPort.Port(),
			"MOCK_FILE_PATH":              "/tmp/modrinth",
			"WHITELISTED_MODPACK_DOMAINS": `["localhost"]`,
		}),
		testcontainers.WithCmd("/labrinth/labrinth", "--no-background-tasks", "--no-migrations"),
		testcontainers.WithMounts(testcontainers.VolumeMount(s.volume, "/tmp/modrinth")),
		testcontainers.WithExposedPorts("8000/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/v2/tag/loader").WithPort("8000/tcp")),
	)
	if err != nil {
		return err
	}
	labrinthURL, err = labrinth.PortEndpoint(ctx, "8000/tcp", "http")
	if err != nil {
		return fmt.Errorf("labrinth endpoint: %w", err)
	}

	return nil
}

// runContainer starts a container on the stack's network, reachable there as alias.
// A container that was created is recorded for teardown even when it
// failed to become ready.
func (s *stack) runContainer(ctx context.Context, alias string, image string, opts ...testcontainers.ContainerCustomizer) (*testcontainers.DockerContainer, error) {
	opts = append(opts, network.WithNetwork([]string{alias}, s.network))
	ctr, err := testcontainers.Run(ctx, image, opts...)
	if ctr != nil {
		s.containers = append(s.containers, ctr)
	}
	if err != nil {
		if ctr != nil {
			return nil, fmt.Errorf("start %s: %w\n%s", alias, err, tail(ctx, ctr))
		}
		return nil, fmt.Errorf("start %s: %w", alias, err)
	}
	return ctr, nil
}

// terminate removes containers in reverse start order, then the volume and
// the network. It is safe to call on a partially started stack.
func (s *stack) terminate() error {
	var errs []error
	for i := len(s.containers) - 1; i >= 0; i-- {
		// Nothing in the stack outlives the run, so a graceful stop
		// only adds the stop timeout per container.
		opts := []testcontainers.TerminateOption{testcontainers.StopTimeout(0)}
		// The volume can only go once no container uses it, which is
		// after the last container is removed.
		if i == 0 {
			opts = append(opts, testcontainers.RemoveVolumes(s.volume))
		}
		errs = append(errs, testcontainers.TerminateContainer(s.containers[i], opts...))
	}
	if s.network != nil {
		errs = append(errs, s.network.Remove(context.Background()))
	}
	return errors.Join(errs...)
}

// tail returns the last lines of a container's log, for error messages.
func tail(ctx context.Context, ctr testcontainers.Container) string {
	const n = 20
	r, err := ctr.Logs(ctx)
	if err != nil {
		return fmt.Sprintf("(no logs: %v)", err)
	}
	defer r.Close()
	var lines []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	if err := sc.Err(); err != nil {
		lines = append(lines, fmt.Sprintf("(log truncated: %v)", err))
	}
	return strings.Join(lines, "\n")
}
