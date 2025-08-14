module github.com/0xDezzy/langchaingo

go 1.23.0

toolchain go1.24.3

// Note: Thanks to Go's module graph pruning (https://go.dev/ref/mod#graph-pruning),
// importing langchaingo does NOT pull in all dependencies listed below. You only
// get dependencies for the specific packages you import. For example:
//   - import "github.com/tmc/langchaingo/llms/openai" → only OpenAI-related deps
//   - import "github.com/tmc/langchaingo/vectorstores/chroma" → only Chroma deps
// This keeps your builds lean despite this large go.mod file.

// Core dependencies
require github.com/google/uuid v1.6.0

// Testing
require (
	github.com/stretchr/testify v1.10.0
	github.com/testcontainers/testcontainers-go v0.37.0
	github.com/testcontainers/testcontainers-go/modules/chroma v0.37.0
	github.com/testcontainers/testcontainers-go/modules/milvus v0.37.0
	github.com/testcontainers/testcontainers-go/modules/mongodb v0.37.0
	github.com/testcontainers/testcontainers-go/modules/mysql v0.37.0
	github.com/testcontainers/testcontainers-go/modules/opensearch v0.37.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.37.0
	github.com/testcontainers/testcontainers-go/modules/redis v0.37.0
	github.com/testcontainers/testcontainers-go/modules/weaviate v0.37.0
)

// LLM providers
require (
	github.com/IBM/watsonx-go v1.0.0
	github.com/cohere-ai/tokenizer v1.1.2
	github.com/gage-technologies/mistral-go v1.1.0
	github.com/google/generative-ai-go v0.15.1
	github.com/pkoukk/tiktoken-go v0.1.6
)

// Vector stores
require (
	github.com/amikos-tech/chroma-go v0.1.4
	github.com/milvus-io/milvus-sdk-go/v2 v2.4.0
	github.com/opensearch-project/opensearch-go v1.1.0
	github.com/pgvector/pgvector-go v0.1.1
	github.com/pinecone-io/go-pinecone v0.4.1
	github.com/redis/rueidis v1.0.34
	github.com/weaviate/weaviate v1.29.0
	github.com/weaviate/weaviate-go-client/v5 v5.0.2
	go.mongodb.org/mongo-driver v1.14.0
	go.mongodb.org/mongo-driver/v2 v2.0.0
)

// Cloud platforms and AI services
require (
	cloud.google.com/go/aiplatform v1.69.0
	cloud.google.com/go/vertexai v0.12.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.4
	github.com/aws/aws-sdk-go-v2/service/bedrockagent v1.40.0
	github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime v1.41.0
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.24.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.78.2
	github.com/aws/smithy-go v1.22.2
	golang.org/x/oauth2 v0.30.0
	google.golang.org/api v0.218.0
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.3
)

// Embeddings and ML
require (
	github.com/AssemblyAI/assemblyai-go-sdk v1.3.0
	github.com/nlpodyssey/cybertron v0.2.1
)

// Database drivers
require (
	cloud.google.com/go/alloydbconn v1.13.2
	cloud.google.com/go/cloudsqlconn v1.14.1
	github.com/go-sql-driver/mysql v1.8.1
	github.com/jackc/pgx/v5 v5.7.2
	github.com/mattn/go-sqlite3 v1.14.17
)

// Document processing and web scraping
require (
	github.com/PuerkitoBio/goquery v1.8.1
	github.com/gocolly/colly v1.2.0
	github.com/ledongthuc/pdf v0.0.0-20220302134840-0c2507a12d80
	github.com/microcosm-cc/bluemonday v1.0.26
	gitlab.com/golang-commonmark/markdown v0.0.0-20211110145824-bf3e522c626a
)

// Memory and agent tools
require (
	github.com/metaphorsystems/metaphor-go v0.0.0-20230816231421-43794c04824e
	go.starlark.net v0.0.0-20230302034142-4b1e35fe2254
)

// Utilities and helpers
require (
	github.com/Code-Hex/go-generics-cache v1.3.1
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/go-openapi/strfmt v0.23.0
	github.com/google/go-cmp v0.7.0
	github.com/nikolalohinski/gonja v1.5.3
	github.com/zeebo/blake3 v0.2.4
	golang.org/x/sync v0.15.0
	golang.org/x/tools v0.33.0
	sigs.k8s.io/yaml v1.3.0
)

// Indirect dependencies (automatically managed)

// Cloud platforms and AI services - indirect
require (
	cloud.google.com/go v0.116.0 // indirect
	cloud.google.com/go/ai v0.7.0 // indirect
	cloud.google.com/go/alloydb v1.14.0 // indirect
	cloud.google.com/go/auth v0.14.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.2.2 // indirect
	cloud.google.com/go/longrunning v0.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.57 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.12 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.54.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250122153221-138b5a5a4fd4 // indirect
)

// Embeddings and ML - indirect
require (
	github.com/nlpodyssey/gopickle v0.2.0 // indirect
	github.com/nlpodyssey/gotokenizers v0.2.0 // indirect
	github.com/nlpodyssey/spago v1.1.0 // indirect
)

// Vector stores - indirect
require (
	github.com/milvus-io/milvus-proto/go-api/v2 v2.4.0 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)

// Database drivers - indirect
require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
)

// Cryptography and security - indirect
require filippo.io/edwards25519 v1.1.0 // indirect

// Document processing and web scraping - indirect
require (
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/antchfx/htmlquery v1.3.0 // indirect
	github.com/antchfx/xmlquery v1.3.17 // indirect
	github.com/antchfx/xpath v1.2.4 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/kennygrant/sanitize v1.2.4 // indirect
	github.com/saintfish/chardet v0.0.0-20230101081208-5e3ef4b5456d // indirect
	github.com/temoto/robotstxt v1.1.2 // indirect
	gitlab.com/golang-commonmark/html v0.0.0-20191124015941-a22733972181 // indirect
	gitlab.com/golang-commonmark/linkify v0.0.0-20191026162114-a0c2df6c8f82 // indirect
	gitlab.com/golang-commonmark/mdurl v0.0.0-20191124015652-932350d1cb84 // indirect
	gitlab.com/golang-commonmark/puny v0.0.0-20191124015043-9f83538fa04f // indirect
)

// Testing infrastructure - indirect
require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v28.2.2+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/getsentry/sentry-go v0.30.0 // indirect
	github.com/mdelapenya/tlscert v0.2.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
)

// Utilities and common libraries - indirect
require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/cockroachdb/errors v1.9.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/deepmap/oapi-codegen/v2 v2.1.0 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/errors v0.22.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/runtime v0.24.2 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-openapi/validate v0.24.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v23.5.26+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/rs/zerolog v1.31.0 // indirect
	github.com/shirou/gopsutil/v4 v4.25.5 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/yargevad/filepathx v1.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/exp v0.0.0-20240808152545-0cdaa3abc0fa // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/FalkorDB/falkordb-go v1.0.0
	github.com/getzep/zep-go/v2 v2.22.0
	github.com/gomodule/redigo v1.7.1-0.20190724094224-574c33c3df38
	github.com/jaytaylor/html2text v0.0.0-20230321000545-74c2419ad056
	github.com/timsims/pamphlet v0.1.6
)

require (
	github.com/anush008/fastembed-go v1.0.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/rivo/uniseg v0.4.6 // indirect
	github.com/schollz/progressbar/v2 v2.15.0 // indirect
	github.com/schollz/progressbar/v3 v3.14.1 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/sugarme/regexpset v0.0.0-20200920021344-4d4ec8eaf93c // indirect
	github.com/sugarme/tokenizer v0.2.3-0.20230829214935-448e79b1ed65 // indirect
	github.com/yalue/onnxruntime_go v1.7.0 // indirect
	golang.org/x/term v0.32.0 // indirect
)
