package documentloaders

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// ErrMissingAudioSource is returned when neither an audio URL nor a reader has
// been set using [WithAudioURL] or [WithAudioReader].
var ErrMissingAudioSource = errors.New("assemblyai: missing audio source")

// TranscriptFormat represents the format of the document page content.
type TranscriptFormat int

const (
	// Single document with full transcript text.
	TranscriptFormatText TranscriptFormat = iota

	// Multiple documents with each sentence as page content.
	TranscriptFormatSentences

	// Multiple documents with each paragraph as page content.
	TranscriptFormatParagraphs

	// Single document with SRT formatted subtitles as page content.
	TranscriptFormatSubtitlesSRT

	// Single document with VTT formatted subtitles as page content.
	TranscriptFormatSubtitlesVTT
)

// AssemblyAIAudioTranscriptLoader transcribes an audio file using AssemblyAI
// and loads the transcript.
//
// Audio files can be specified using either a URL or a reader.
//
// For a list of the supported audio and video formats, see the [FAQ].
//
// [FAQ]: https://www.assemblyai.com/docs/concepts/faq
type AssemblyAIAudioTranscriptLoader struct {
	client *assemblyai.Client

	// URL of the audio file to transcribe.
	url string

	// Reader of the audio file to transcribe.
	r io.Reader

	// Optional parameters for the transcription.
	params *assemblyai.TranscriptOptionalParams

	// Format of the document page content.
	format TranscriptFormat
}

var _ Loader = &AssemblyAIAudioTranscriptLoader{}

// AssemblyAIOption is an option for the AssemblyAI loader.
type AssemblyAIOption func(loader *AssemblyAIAudioTranscriptLoader)

// NewAssemblyAIAudioTranscript returns a new instance
// AssemblyAIAudioTranscriptLoader.
func NewAssemblyAIAudioTranscript(apiKey string, opts ...AssemblyAIOption) *AssemblyAIAudioTranscriptLoader {
	loader := &AssemblyAIAudioTranscriptLoader{
		client: assemblyai.NewClient(apiKey),
		format: TranscriptFormatText,
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader
}

// WithAudioURL configures the loader to transcribe an audio file from a URL.
// The URL needs to be accessible from AssemblyAI's servers.
func WithAudioURL(url string) AssemblyAIOption {
	return func(a *AssemblyAIAudioTranscriptLoader) {
		a.url = url
	}
}

// WithAudioReader configures the loader to transcribe a local audio file.
func WithAudioReader(r io.Reader) AssemblyAIOption {
	return func(a *AssemblyAIAudioTranscriptLoader) {
		a.r = r
	}
}

// WithAudioReader configures the format of the document page content.
func WithTranscriptFormat(format TranscriptFormat) AssemblyAIOption {
	return func(a *AssemblyAIAudioTranscriptLoader) {
		a.format = format
	}
}

// WithTranscriptParams configures the optional parameters for the transcription.
func WithTranscriptParams(params *assemblyai.TranscriptOptionalParams) AssemblyAIOption {
	return func(a *AssemblyAIAudioTranscriptLoader) {
		a.params = params
	}
}

// Load transcribes an audio file, transcribes it using AssemblyAI, and returns
// them transcript as a document.
func (a *AssemblyAIAudioTranscriptLoader) Load(ctx context.Context) ([]schema.Document, error) {
	transcript, err := a.transcribe(ctx)
	if err != nil {
		return nil, err
	}

	docs, err := a.formatTranscript(ctx, transcript)
	if err != nil {
		return nil, err
	}

	return docs, nil
}

// transcribe conditionally transcribes an audio file based on the specified
// source.
func (a *AssemblyAIAudioTranscriptLoader) transcribe(ctx context.Context) (assemblyai.Transcript, error) {
	if a.url != "" {
		return a.client.Transcripts.TranscribeFromURL(ctx, a.url, a.params)
	}

	if a.r != nil {
		return a.client.Transcripts.TranscribeFromReader(ctx, a.r, a.params)
	}

	return assemblyai.Transcript{}, ErrMissingAudioSource
}

// formatTranscript returns a schema.Document for a transcript based on the
// specific format.
func (a *AssemblyAIAudioTranscriptLoader) formatTranscript(ctx context.Context, transcript assemblyai.Transcript) ([]schema.Document, error) {
	switch a.format {
	case TranscriptFormatSentences:
		sentences, err := a.client.Transcripts.GetSentences(ctx, assemblyai.ToString(transcript.ID))
		if err != nil {
			return nil, err
		}
		return documentsFromSentences(sentences.Sentences)

	case TranscriptFormatParagraphs:
		paragraphs, err := a.client.Transcripts.GetParagraphs(ctx, assemblyai.ToString(transcript.ID))
		if err != nil {
			return nil, err
		}
		return documentsFromParagraphs(paragraphs.Paragraphs)

	case TranscriptFormatSubtitlesSRT:
		srt, err := a.client.Transcripts.GetSubtitles(ctx, assemblyai.ToString(transcript.ID), "srt", nil)
		if err != nil {
			return nil, err
		}
		return []schema.Document{{PageContent: string(srt)}}, nil

	case TranscriptFormatSubtitlesVTT:
		vtt, err := a.client.Transcripts.GetSubtitles(ctx, assemblyai.ToString(transcript.ID), "vtt", nil)
		if err != nil {
			return nil, err
		}
		return []schema.Document{{PageContent: string(vtt)}}, nil

	case TranscriptFormatText:
		fallthrough

	default:
		metadata, err := toMetadata(transcript)
		if err != nil {
			return nil, err
		}
		return []schema.Document{{PageContent: assemblyai.ToString(transcript.Text), Metadata: metadata}}, nil
	}
}

func documentsFromSentences(sentences []assemblyai.TranscriptSentence) ([]schema.Document, error) {
	docs := make([]schema.Document, 0, len(sentences))

	for _, sentence := range sentences {
		metadata, err := toMetadata(sentence)
		if err != nil {
			return nil, err
		}

		docs = append(docs, schema.Document{
			PageContent: assemblyai.ToString(sentence.Text),
			Metadata:    metadata,
		})
	}

	return docs, nil
}

func documentsFromParagraphs(paragraphs []assemblyai.TranscriptParagraph) ([]schema.Document, error) {
	docs := make([]schema.Document, 0, len(paragraphs))

	for _, paragraph := range paragraphs {
		metadata, err := toMetadata(paragraph)
		if err != nil {
			return nil, err
		}

		docs = append(docs, schema.Document{
			PageContent: assemblyai.ToString(paragraph.Text),
			Metadata:    metadata,
		})
	}

	return docs, nil
}

// toMetadata converts a struct to a map representation to use as metadata.
func toMetadata(obj any) (map[string]any, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var metadata map[string]any
	if err := json.Unmarshal(b, &metadata); err != nil {
		return nil, err
	}

	// Remove redundant transcript text.
	delete(metadata, "text")

	return metadata, nil
}

// LoadAndSplit transcribes the audio data and splits it into multiple documents
// using a text splitter.
func (a *AssemblyAIAudioTranscriptLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := a.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}
