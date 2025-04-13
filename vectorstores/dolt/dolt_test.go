package dolt_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tmc/langchaingo/vectorstores"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/dolt"
)

var (
	//nolint:gochecknoglobals
	doltExec string
	//nolint:gochecknoglobals
	doltExecOnce sync.Once
)

type testDoltServer struct {
	t              *testing.T
	Cmd            *exec.Cmd
	db             *sql.DB
	Stdout         io.ReadCloser
	Stderr         io.ReadCloser
	Name           string
	StderrString   string
	StderrCaptured chan (bool)
	WaitError      error
	Waited         chan (bool)
	CmdDir         string
	Host           string
	Port           string
	Password       string
}

func newTestDoltServer(t *testing.T) *testDoltServer {
	t.Helper()
	return &testDoltServer{
		t:              t,
		Waited:         make(chan bool),
		StderrCaptured: make(chan bool),
		Name:           "vectorstore_dolt_test",
	}
}

func mustGetDoltExec(t *testing.T) string {
	t.Helper()

	doltCommand := "dolt"
	if runtime.GOOS == "windows" {
		doltCommand = "dolt.exe"
	}

	doltExecOnce.Do(func() {
		arg := os.Getenv("DOLT_BIN")
		if arg != "" {
			if filepath.IsAbs(arg) {
				doltExec = arg
				return
			}
			wd, _ := os.Getwd()
			doltExec = filepath.Join(wd, arg)
			return
		}
		de, err := exec.LookPath(doltCommand)
		if err != nil {
			t.Skip("Dolt binary not available")
		}
		doltExec = de
	})
	return doltExec
}

func (di *testDoltServer) ConnectionString() string {
	return fmt.Sprintf("%s:%s@(%s:%s)/%s?parseTime=true&multiStatements=true", "root", di.Password, di.Host, di.Port, di.Name)
}

//nolint:funlen
func (di *testDoltServer) Start() error {
	tmpDir, err := os.MkdirTemp("", "dolt-vectorstore-tests*")
	require.NoError(di.t, err)

	di.CmdDir = tmpDir

	doltInit := exec.Command(mustGetDoltExec(di.t), "init") //nolint:gosec
	doltInit.Env = os.Environ()
	doltInit.Dir = tmpDir
	doltInit.Stdout = os.Stdout
	doltInit.Stderr = os.Stderr
	err = doltInit.Run()
	require.NoError(di.t, err)

	createDB := exec.Command(mustGetDoltExec(di.t), "sql", "-q", fmt.Sprintf("CREATE DATABASE %s;", di.Name)) //nolint:gosec
	createDB.Env = os.Environ()
	createDB.Dir = tmpDir
	createDB.Stdout = os.Stdout
	createDB.Stderr = os.Stderr
	err = createDB.Run()
	require.NoError(di.t, err)

	port, err := getFreePort()
	require.NoError(di.t, err)

	di.Host = "0.0.0.0"
	di.Port = port
	di.Password = ""

	di.Cmd = exec.Command( //nolint:gosec
		mustGetDoltExec(di.t),
		"sql-server",
		"--host", di.Host,
		"--port", di.Port,
	)

	di.Cmd.Env = di.Cmd.Environ()
	di.Cmd.Dir = di.CmdDir

	di.Stdout, err = di.Cmd.StdoutPipe()
	require.NoError(di.t, err)
	di.Stderr, err = di.Cmd.StderrPipe()
	require.NoError(di.t, err)

	err = di.Cmd.Start()
	require.NoError(di.t, err)
	go func() {
		di.WaitError = di.Cmd.Wait()
		close(di.Waited)
	}()

	go func() {
		var buffer bytes.Buffer
		_, err := buffer.ReadFrom(di.Stderr)
		if err != nil {
			panic(err)
		}
		di.StderrString = buffer.String()
		close(di.StderrCaptured)
	}()

	dbChan := make(chan *sql.DB)
	go func() {
		for i := 0; i < 50; i++ {
			db, err := sql.Open("mysql", di.ConnectionString())
			if err == nil {
				err = db.Ping()
				if err == nil {
					dbChan <- db
					return
				}
			}
			select {
			case <-di.Waited:
				close(dbChan)
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
		err = di.Shutdown()
		if err != nil {
			panic(err)
		}
		close(dbChan)
	}()
	di.db = <-dbChan

	return nil
}

func (di *testDoltServer) IsRunning() bool {
	return di.Cmd.Process != nil && di.Cmd.ProcessState == nil && di.db != nil && di.db.Ping() == nil
}

func (di *testDoltServer) Shutdown() error {
	defer os.RemoveAll(di.CmdDir)

	killed := false
	if runtime.GOOS == "windows" {
		kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(di.Cmd.Process.Pid)) //nolint:gosec
		kill.Stdout = os.Stdout
		kill.Stderr = os.Stderr
		err := kill.Run()
		if err != nil {
			return err
		}
		killed = true
	} else {
		err := di.Cmd.Process.Signal(os.Interrupt)
		if err != nil {
			return err
		}
	}
	<-di.Waited
	<-di.StderrCaptured
	if killed && di.WaitError != nil {
		return nil
	}
	return di.WaitError
}

func (di *testDoltServer) ErrorMessage() string {
	return di.StderrString
}

func (di *testDoltServer) DB() (*sql.DB, error) {
	if !di.IsRunning() {
		return nil, errors.New("dolt server is not running")
	}
	return di.db, nil
}

func getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return "", errors.New("failed to get port")
	}
	return fmt.Sprintf("%d", addr.Port), nil
}

func preCheckEnvSetting(t *testing.T) string {
	t.Helper()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	doltURL := os.Getenv("DOLT_CONNECTION_STRING")
	if doltURL == "" {
		di := newTestDoltServer(t)
		err := di.Start()
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, di.Shutdown())
		})
		doltURL = di.ConnectionString()
	}

	return doltURL
}

func makeNewDatabaseName() string {
	return fmt.Sprintf("test-database-%s", uuid.New().String())
}

func cleanupTestArtifacts(ctx context.Context, t *testing.T, s dolt.Store, doltURL string) {
	t.Helper()

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	require.NoError(t, s.RemoveDatabase(ctx, tx))

	require.NoError(t, tx.Commit())
}

func TestDoltStoreRest(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}

func TestDoltStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London"},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0.6), // Dolt uses euclidean squared distance
	)
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestDoltStoreSimilarityScore(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Tokyo is the capital city of Japan."},
		{PageContent: "Paris is the city of love."},
		{PageContent: "I like to visit London."},
	})
	require.NoError(t, err)

	// Dolt uses euclidean squared distance
	// test with a score threshold of 0.6, expected 6 documents
	docs, err := store.SimilaritySearch(
		ctx,
		"What is the capital city of Japan?",
		3,
		vectorstores.WithScoreThreshold(0.6),
	)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.True(t, docs[0].Score > 0.8)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London"},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	_, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(-0.8),
	)
	require.Error(t, err)

	_, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(1.8),
	)
	require.Error(t, err)
}

// note, we can also use same llm to show this test, but need imply
// openai embedding [dimensions](https://platform.openai.com/docs/api-reference/embeddings/create#embeddings-create-dimensions) args.
func TestSimilaritySearchWithDifferentDimensions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	doltURL := preCheckEnvSetting(t)
	genaiKey := os.Getenv("GENAI_API_KEY")
	if genaiKey == "" {
		t.Skip("GENAI_API_KEY not set")
	}
	databaseName := makeNewDatabaseName()

	// use Google embedding (now default model is embedding-001, with dimensions:768) to add some data to collection
	googleLLM, err := googleai.New(ctx, googleai.WithAPIKey(genaiKey))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(googleLLM)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(databaseName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Beijing"},
	})
	require.NoError(t, err)

	// use openai embedding (now default model is text-embedding-ada-002, with dimensions:1536) to add some data to same collection (same table)
	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err = embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err = dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(false),
		dolt.WithDatabaseName(databaseName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London"},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		5,
	)
	require.NoError(t, err)
	require.Len(t, docs, 5)
}

func TestDoltAsRetriever(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 1),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}

func TestDoltAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithScoreThreshold(0.7)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestDoltAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{
				PageContent: "in kitchen, The color of the lamp beside the desk is black.",
				Metadata: map[string]any{
					"location": "kitchen",
				},
			},
			{
				PageContent: "in bedroom, The color of the lamp beside the desk is blue.",
				Metadata: map[string]any{
					"location": "bedroom",
				},
			},
			{
				PageContent: "in office, The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location": "office",
				},
			},
			{
				PageContent: "in sitting room, The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location": "sitting room",
				},
			},
			{
				PageContent: "in patio, The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location": "patio",
				},
			},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	result = strings.ToLower(result)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "blue", "expected blue in result")
	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "purple", "expected purple in result")
	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestDoltAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{
				PageContent: "In office, the color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location":    "office",
					"square_feet": 100,
				},
			},
			{
				PageContent: "in sitting room, the color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location":    "sitting room",
					"square_feet": 400,
				},
			},
			{
				PageContent: "in patio, the color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location":    "patio",
					"square_feet": 800,
				},
			},
		},
	)
	require.NoError(t, err)

	filter := map[string]any{"location": "sitting room"}

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store,
				5,
				vectorstores.WithFilters(filter))),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	require.Contains(t, result, "purple", "expected purple in result")
	require.NotContains(t, result, "orange", "expected not orange in result")
	require.NotContains(t, result, "yellow", "expected not yellow in result")
}

func TestDeduplicater(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"type": "city",
		}},
		{PageContent: "potato", Metadata: map[string]any{
			"type": "vegetable",
		}},
	}, vectorstores.WithDeduplicater(
		func(_ context.Context, doc schema.Document) bool {
			return doc.PageContent == "tokyo"
		},
	))
	require.NoError(t, err)

	docs, err := store.Search(ctx, 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "potato", docs[0].PageContent)
	require.Equal(t, "vegetable", docs[0].Metadata["type"])
}

func TestWithAllOptions(t *testing.T) {
	t.Parallel()
	doltURL := preCheckEnvSetting(t)
	ctx := context.Background()

	llm, err := openai.New(
		openai.WithEmbeddingModel("text-embedding-ada-002"),
	)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	require.NoError(t, err)
	db, err := sql.Open("mysql", doltURL)
	require.NoError(t, err)
	defer db.Close()

	store, err := dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
		dolt.WithCollectionTableName("collection_table_name"),
		dolt.WithEmbeddingTableName("embedding_table_name"),
		dolt.WithDatabaseMetadata(map[string]any{
			"key": "value",
		}),
		dolt.WithVectorDimensions(1536),
		dolt.WithCreateEmbeddingIndexAfterAddDocuments(true),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])

	store, err = dolt.New(
		ctx,
		dolt.WithDB(db),
		dolt.WithEmbedder(e),
		dolt.WithPreDeleteDatabase(true),
		dolt.WithDatabaseName(makeNewDatabaseName()),
		dolt.WithCollectionTableName("collection_table_name1"),
		dolt.WithEmbeddingTableName("embedding_table_name1"),
		dolt.WithDatabaseMetadata(map[string]any{
			"key": "value",
		}),
		dolt.WithVectorDimensions(1536),
		dolt.WithCreateEmbeddingIndexAfterAddDocuments(true),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, doltURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err = store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}
