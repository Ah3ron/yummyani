package api_test

import (
	"encoding/json"
	"testing"

	"yummyani/internal/api"
)

func TestFlexInt_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"integer", `42`, 42, false},
		{"string integer", `"42"`, 42, false},
		{"string with spaces", `" 42 "`, 42, false},
		{"zero", `0`, 0, false},
		{"negative", `-5`, -5, false},
		{"negative string", `"-5"`, -5, false},
		{"invalid string", `"abc"`, 0, true},
		{"empty", ``, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fi api.FlexInt
			err := json.Unmarshal([]byte(tt.input), &fi)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlexInt.UnmarshalJSON(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && int(fi) != tt.want {
				t.Errorf("FlexInt.UnmarshalJSON(%q) = %d, want %d", tt.input, int(fi), tt.want)
			}
		})
	}
}

func TestGroupByDubbing(t *testing.T) {
	tests := []struct {
		name   string
		videos []api.VideoEntry
		want   int // number of groups
	}{
		{
			name:   "empty",
			videos: nil,
			want:   0,
		},
		{
			name: "single kodik group",
			videos: []api.VideoEntry{
				{Data: api.VideoData{Dubbing: "Studio A", Player: "Kodik Player"}},
				{Data: api.VideoData{Dubbing: "Studio A", Player: "Kodik Player"}},
			},
			want: 1,
		},
		{
			name: "two kodik groups",
			videos: []api.VideoEntry{
				{Data: api.VideoData{Dubbing: "Studio A", Player: "Kodik Player"}},
				{Data: api.VideoData{Dubbing: "Studio B", Player: "Kodik Player"}},
			},
			want: 2,
		},
		{
			name: "non-kodik filtered out",
			videos: []api.VideoEntry{
				{Data: api.VideoData{Dubbing: "Studio A", Player: "OtherPlayer"}},
				{Data: api.VideoData{Dubbing: "Studio B", Player: "Kodik Player"}},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := api.GroupByDubbing(tt.videos)
			if len(got) != tt.want {
				t.Errorf("GroupByDubbing() got %d groups, want %d", len(got), tt.want)
			}
		})
	}
}

func TestAnimeType_DisplayName(t *testing.T) {
	tests := []struct {
		name string
		t    api.AnimeType
		want string
	}{
		{"with shortname", api.AnimeType{Shortname: "TV", Name: "Television"}, "TV"},
		{"without shortname", api.AnimeType{Shortname: "", Name: "Movie"}, "Movie"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.DisplayName(); got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAnimeInfo_DisplayName(t *testing.T) {
	tests := []struct {
		name string
		a    api.AnimeInfo
		want string
	}{
		{"title优先", api.AnimeInfo{Title: "Naruto", Name: "Наурото"}, "Naruto"},
		{"fallback name", api.AnimeInfo{Title: "", Name: "Наурото"}, "Наурото"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.DisplayName(); got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}
