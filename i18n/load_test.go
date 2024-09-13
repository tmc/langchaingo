package i18n

import (
	"strings"
	"testing"
)

func Test_mustLoad(t *testing.T) {
	t.Parallel()
	type args struct {
		lang       Lang
		kindFolder string
		filename   string
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
				lang:       ZH,
				kindFolder: "agents",
				filename:   "mrkl_prompt_format_instructions.txt",
			},
			want: `严格遵循以下格式:

问题: 你要回答的问题
思考: 针对这个问题，你的思考过程
工具: 你将要使用的工具，只能在 [ {{.tool_names}} ] 中选择一个
工具参数: 使用工具所需的参数
结果: 使用工具获得的结果
...（『思考/工具/工具参数/结果』可以重复多次）
思考: 我知道最终答案了
最终答案: 问题的最终答案`,
		},
		{
			name: "should panic due to unknown language",
			args: args{
				lang:       Lang(-1),
				kindFolder: "whatever",
				filename:   "whatever",
			},
			wantPanic:     true,
			wantPanicLike: "unknown language",
		},
		{
			name: "should panic due to reading file failed",
			args: args{
				lang:       EN,
				kindFolder: "agents",
				filename:   "whatever",
			},
			wantPanic:     true,
			wantPanicLike: "read file failed",
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

			if got := mustLoad(tt.args.lang, tt.args.kindFolder, tt.args.filename); got != tt.want {
				t.Errorf("mustLoad() = %v, want %v", got, tt.want)
			}
		})
	}
}
