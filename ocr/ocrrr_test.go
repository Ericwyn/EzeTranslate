package ocr

import "testing"

func TestSelectTesseractLanguage(t *testing.T) {
	tests := []struct {
		name      string
		available map[string]bool
		want      string
		wantErr   bool
	}{
		{name: "both languages", available: map[string]bool{"eng": true, "chi_sim": true}, want: "chi_sim+eng"},
		{name: "english only", available: map[string]bool{"eng": true}, want: "eng"},
		{name: "simplified chinese only", available: map[string]bool{"chi_sim": true}, want: "chi_sim"},
		{name: "no supported language", available: map[string]bool{"osd": true}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectTesseractLanguage(tt.available)
			if (err != nil) != tt.wantErr {
				t.Fatalf("selectTesseractLanguage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("selectTesseractLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}
