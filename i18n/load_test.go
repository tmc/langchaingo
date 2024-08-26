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
			want: `请严格遵循以下格式:

问题: 你要回答的问题
思考: 针对这个问题，你的思考过程
动作: 你即将要采取的动作或使用的工具，必须是 [ 计算器, 谷歌搜索 ] 中的一个
动作参数: 你要采取的动作或使用的工具所需的参数
动作结果: 这次动作取得的结果
...（这套『问题/思考/动作/动作参数/动作结果』的格式可以重复多次）
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
