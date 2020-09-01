/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
