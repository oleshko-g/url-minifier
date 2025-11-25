package http

import "testing"

func Test_parseCoding(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		s       string
		want    parsedCoding
		wantErr bool
	}{
		{
			name:    "quality value has more then 3 digits after the point",
			s:       "*;q=1.0000",
			want:    parsedCoding{},
			wantErr: true,
		},
		{
			name:    "quality value is infinity",
			s:       "*;q=infinity",
			want:    parsedCoding{},
			wantErr: true,
		},
		{
			name: "quality value 1.000",
			s:    "*;q=1.000",
			want: parsedCoding{
				coding:       "*",
				qualityValue: 1.000,
			},
			wantErr: false,
		},
		{
			name: "quality value 0.900",
			s:    "*;q=0.900",
			want: parsedCoding{
				coding:       "*",
				qualityValue: 0.9,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseCoding(tt.s)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseCoding() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseCoding() succeeded unexpectedly")
			}
			if tt.want != got {
				t.Errorf("parseCoding() = %v, want %v", got, tt.want)
			}
		})
	}
}
