package i18n

import (
	"strings"
	"testing"
)

func TestAgentsMustPhrase(t *testing.T) {
	t.Parallel()
	type args struct {
		lang Lang
		key  string
	}
	tests := []struct {
		name          string
		args          args
		want          string
		wantPanic     bool
		wantPanicLike string
	}{
		{
			name: "should succeed",
			args: args{
				lang: ZH,
				key:  "thought",
			},
			want: "思考:",
		},
		{
			name: "should panic",
			args: args{
				lang: ZH,
				key:  "invalid key",
			},
			wantPanic:     true,
			wantPanicLike: "there is no such key in phrase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("want panic, but did not happen")
					} else if s, ok := r.(string); !ok {
						t.Errorf("unexpected panic type")
					} else if !strings.Contains(s, tt.wantPanicLike) {
						t.Errorf("panic = %v, want %v", r, tt.wantPanicLike)
					}
				}()
			}

			if got := AgentsMustPhrase(tt.args.lang, tt.args.key); got != tt.want {
				t.Errorf("AgentsMustPhrase() = %v, want %v", got, tt.want)
			}
		})
	}
}
