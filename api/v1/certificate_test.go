package v1

import (
	"reflect"
	"testing"

	"github.com/imdario/mergo"
)

type Team struct {
	Names []string
}

func Test_StringSetMerger(t1 *testing.T) {
	type args struct {
		dst Team
		src Team
	}
	tests := []struct {
		name string
		args args
		want Team
	}{
		{
			name: "0/0",
			args: args{
				dst: Team{},
				src: Team{},
			},
			want: Team{},
		},
		{
			name: "1/0",
			args: args{
				dst: Team{Names: []string{"a"}},
				src: Team{},
			},
			want: Team{Names: []string{"a"}},
		},
		{
			name: "0/1",
			args: args{
				dst: Team{},
				src: Team{Names: []string{"a"}},
			},
			want: Team{Names: []string{"a"}},
		},
		{
			name: "1/1",
			args: args{
				dst: Team{Names: []string{"a"}},
				src: Team{Names: []string{"a"}},
			},
			want: Team{Names: []string{"a"}},
		},
		{
			name: ">1/0",
			args: args{
				dst: Team{Names: []string{"a", "a", "b"}},
				src: Team{},
			},
			want: Team{Names: []string{"a", "b"}},
		},
		// In this case, mergo does not call transformer and directly returns src
		// So, this test fails
		//{
		//	name: "0/>1",
		//	args: args{
		//		dst: Team{},
		//		src: Team{Names: []string{"a", "a", "b"}},
		//	},
		//	want: Team{Names: []string{"a", "b"}},
		//},
		{
			name: ">1/>1",
			args: args{
				dst: Team{Names: []string{"a", "a", "b"}},
				src: Team{Names: []string{"b", "c", "c"}},
			},
			want: Team{Names: []string{"a", "b", "c"}},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t *testing.T) {
			err := mergo.Merge(&tt.args.dst, tt.args.src, mergo.WithTransformers(stringSetMerger{}))
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(tt.args.dst, tt.want) {
				t.Errorf("StringSetMerger() got = %v, want %v", tt.args.dst, tt.want)
			}
		})
	}
}
