package queryconstructor_prompts

import _ "embed"

//go:embed example_full_answer.txt
var _exampleFullAnswer string //nolint:gochecknoglobals

//go:embed example_no_filter_answer.txt
var _exampleNoFilterAnswer string //nolint:gochecknoglobals

//go:embed example_song_data_source.txt
var _exampleSongDataSource string //nolint:gochecknoglobals

//go:embed example_with_limit_answer.txt
var _exampleWithLimitAnswer string //nolint:gochecknoglobals

var (
	DefaultExamples []map[string]string = []map[string]string{
		{
			"i":                  "1",
			"data_source":        _exampleSongDataSource,
			"user_query":         "What are songs by Taylor Swift or Katy Perry about teenage romance under 3 minutes long in the dance pop genre",
			"structured_request": _exampleFullAnswer,
		},
		{
			"i":                  "2",
			"data_source":        _exampleSongDataSource,
			"user_query":         "What are songs that were not published on Spotify",
			"structured_request": _exampleNoFilterAnswer,
		},
	}

	ExamplesWithLimit []map[string]string = []map[string]string{
		{
			"i":                  "1",
			"data_source":        _exampleSongDataSource,
			"user_query":         "What are songs by Taylor Swift or Katy Perry about teenage romance under 3 minutes long in the dance pop genre",
			"structured_request": _exampleFullAnswer,
		},
		{
			"i":                  "2",
			"data_source":        _exampleSongDataSource,
			"user_query":         "What are songs that were not published on Spotify",
			"structured_request": _exampleNoFilterAnswer,
		},
		{
			"i":                  "3",
			"data_source":        _exampleSongDataSource,
			"user_query":         "What are three songs about love",
			"structured_request": _exampleWithLimitAnswer,
		},
	}
)
