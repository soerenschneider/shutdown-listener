package internal

import "testing"

func Test_runHook(t *testing.T) {
	type args struct {
		command []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "empty",
			args:    args{command: []string{}},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    args{command: []string{"echo"}},
			wantErr: false,
		},
		{
			name:    "arg",
			args:    args{command: []string{"echo", "hello"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runHook(tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("runHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
