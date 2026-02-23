package metrics

import (
	"testing"
)

func Test_GetRuntimeMetric(t *testing.T) {
	tests := []struct {
		testName string
		name     string
		wantErr  bool
	}{
		{
			testName: "Existed metric",
			name:     "Frees",
			wantErr:  false,
		},
		{
			testName: "Nonexistent metric",
			name:     "JustRandomString",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, gotErr := GetRuntimeMetric(tt.name)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetRuntimeMetric() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetRuntimeMetric() succeeded unexpectedly")
			}
		})
	}
}
