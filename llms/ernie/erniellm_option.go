package ernie

const (
	ernieAPIKey    = "ERNIE_API_KEY"    //nolint:gosec
	ernieSecretKey = "ERNIE_SECRET_KEY" //nolint:gosec
)

type ModelName string

const (
	ModelNameERNIEBot       = "ERNIE-Bot"
	ModelNameERNIEBotTurbo  = "ERNIE-Bot-turbo"
	ModelNameBloomz7B       = "BLOOMZ-7B"
	ModelNameLlama2_7BChat  = "Llama-2-7b-chat"
	ModelNameLlama2_13BChat = "Llama-2-13b-chat"
	ModelNameLlama2_70BChat = "Llama-2-70b-chat"
)

type options struct {
	apiKey      string
	secretKey   string
	accessToken string
	modelName   ModelName
}

type Option func(*options)

// WithAKSK passes the ERNIE API Key and Secret Key to the client. If not set, the keys
// are read from the ERNIE_API_KEY and ERNIE_SECRET_KEY environment variable.
// eg:
//
//	export ERNIE_API_KEY={Api Key}
//	export ERNIE_SECRET_KEY={Serect Key}
//
// Api Key,Serect Key from https://console.bce.baidu.com/qianfan/ais/console/applicationConsole/application
// More information available: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/flfmc9do2
func WithAKSK(apiKey, secretKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
		opts.secretKey = secretKey
	}
}

// WithAccessToken usually used for dev, Prod env recommend use WithAKSK.
func WithAccessToken(accessToken string) Option {
	return func(opts *options) {
		opts.accessToken = accessToken
	}
}

// WithModelName passes the Model Name to the client. If not set, use default ERNIE-Bot.
func WithModelName(modelName ModelName) Option {
	return func(opts *options) {
		opts.modelName = modelName
	}
}
