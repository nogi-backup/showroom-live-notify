package main

import "testing"

func Test_extraTitle(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name:  "Test Title",
			args:  args{title: "黒見 明香（乃木坂46）"},
			want:  "黒見明香",
			want1: "乃木坂46",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := extraTitle(tt.args.title)
			if got != tt.want {
				t.Errorf("extraTitle() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("extraTitle() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
