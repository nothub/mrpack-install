//go:build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// labrinthURL is the base URL of the labrinth API, without a trailing slash.
var labrinthURL string

// binary is the path of the mrpack-install binary under test.
var binary string

// cacheHome is the XDG_CACHE_HOME every run of the binary gets, so archive
// caching stays out of the user's own cache.
var cacheHome string

// runTimeout bounds one run of the binary. The slowest run downloads the
// Mojang server jar.
const runTimeout = 5 * time.Minute

// mrpackInstall runs the binary under test with args and returns its
// combined output and exit error. A run that outlives runTimeout fails t.
func mrpackInstall(t testing.TB, args ...string) ([]byte, error) {
	t.Helper()
	var out bytes.Buffer
	err := runBinary(t, &out, &out, args)
	return out.Bytes(), err
}

// mrpackInstallStdout runs the binary like mrpackInstall and returns its
// stdout and stderr apart.
func mrpackInstallStdout(t testing.TB, args ...string) (stdout []byte, stderr []byte, err error) {
	t.Helper()
	var so, se bytes.Buffer
	err = runBinary(t, &so, &se, args)
	return so.Bytes(), se.Bytes(), err
}

func runBinary(t testing.TB, stdout io.Writer, stderr io.Writer, args []string) error {
	t.Helper()
	// No run may read stdin. A stdin that stays open and never delivers
	// data turns a stray read into a hang that hits the timeout, where
	// the default /dev/null would return EOF and let it pass. An io.Pipe
	// would make Wait block on its copying goroutine instead.
	stdin, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	defer w.Close()

	ctx, cancel := context.WithTimeout(t.Context(), runTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cacheHome)
	err = cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("mrpack-install %s: no exit after %s, did it read stdin?", strings.Join(args, " "), runTimeout)
	}
	return err
}

// PATs for the seed dump's regular and admin user, sent as the
// Authorization header. labrinth routes a token to its PAT lookup by the
// mrp_ prefix. The dump's own mra_ sessions for these users expire in 2030.
const (
	userPAT  = "mrp_e2euser"
	adminPAT = "mrp_e2eadmin"
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "mrpack-install-e2e-")
	if err != nil {
		log.Printf("create bin dir: %v", err)
		return 1
	}
	defer os.RemoveAll(dir)
	binary = filepath.Join(dir, "mrpack-install")
	cacheHome = filepath.Join(dir, "cache")
	// Built before the stack starts, so a compile error fails fast.
	build := exec.Command("go", "build", "-o", binary, "github.com/nothub/mrpack-install")
	if out, err := build.CombinedOutput(); err != nil {
		log.Printf("build mrpack-install: %v\n%s", err, out)
		return 1
	}

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

	postgres, err := s.runContainer(ctx, "postgres", imagePostgres,
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
	// A new volume belongs to root, and labrinth runs as whichever user its
	// image picks, so the cdn opens the volume to every user before
	// labrinth starts. Its port wait orders the two.
	cdn, err := s.runContainer(ctx, "cdn", imageBusybox,
		testcontainers.WithCmd("sh", "-c", "chmod 1777 /srv && exec httpd -f -p 80 -h /srv"),
		testcontainers.WithMounts(testcontainers.ContainerMount{
			Source: testcontainers.GenericVolumeMountSource{Name: s.volume},
			Target: "/srv",
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

	if err := seed(ctx, postgres); err != nil {
		return err
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

// seed loads labrinth's seed dump and inserts the test PATs.
func seed(ctx context.Context, postgres testcontainers.Container) error {
	const dump = "/tmp/seed.sql"
	err := postgres.CopyFileToContainer(ctx, filepath.Join("testdata", seedDump), dump, 0o644)
	if err != nil {
		return fmt.Errorf("copy seed dump: %w", err)
	}
	pats := fmt.Sprintf(`INSERT INTO pats (id, name, access_token, scopes, user_id, expires) VALUES
		(1, 'e2e user', '%s', %d, %d, '2100-01-01'),
		(2, 'e2e admin', '%s', %d, %d, '2100-01-01')`,
		userPAT, allScopes, seedUserID,
		adminPAT, allScopes, seedAdminID)
	code, out, err := postgres.Exec(ctx, []string{"psql", "-q",
		"-v", "ON_ERROR_STOP=1", "-U", "labrinth", "-d", "labrinth",
		"-f", dump, "-c", pats,
	}, tcexec.Multiplexed())
	if err != nil {
		return fmt.Errorf("exec psql: %w", err)
	}
	if code != 0 {
		b, err := io.ReadAll(out)
		if err != nil {
			b = fmt.Appendf(b, "(output truncated: %v)", err)
		}
		return fmt.Errorf("seed exited with %d:\n%s", code, b)
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
