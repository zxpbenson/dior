package option

import (
	"os"
	"testing"
)

func TestOptions_Validate(t *testing.T) {
	// Create a temporary file for testing src-file checks
	tmpFile, err := os.CreateTemp("", "dior-test-src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	// Write some content to make it not empty
	if _, err := tmpFile.Write([]byte("some data")); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Create a temporary file that DOES exist to fail dst-file check
	existingDstFile, err := os.CreateTemp("", "dior-test-dst-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(existingDstFile.Name())
	existingDstFile.Close()

	tests := []struct {
		name    string
		opts    *Options
		wantErr bool
	}{
		// --- Source Validations ---
		{
			name: "Valid Kafka Source",
			opts: &Options{
				Src:                 "kafka",
				SrcBootstrapServers: []string{"localhost:9092"},
				SrcGroup:            "group1",
				SrcTopic:            "topic1",
				Dst:                 "nil",
				ChanSize:            100,
			},
			wantErr: false,
		},
		{
			name: "Invalid Kafka Source - Missing Bootstrap",
			opts: &Options{
				Src:      "kafka",
				SrcGroup: "group1",
				SrcTopic: "topic1",
				Dst:      "nil",
			},
			wantErr: true,
		},
		{
			name: "Valid NSQ Source",
			opts: &Options{
				Src:                    "nsq",
				SrcLookupdTCPAddresses: []string{"localhost:4161"},
				SrcChannel:             "ch1",
				SrcTopic:               "topic1",
				Dst:                    "nil",
			},
			wantErr: false,
		},
		{
			name: "Invalid NSQ Source - Missing Addrs",
			opts: &Options{
				Src:        "nsq",
				SrcChannel: "ch1",
				SrcTopic:   "topic1",
				Dst:        "nil",
			},
			wantErr: true,
		},
		{
			name: "Valid Press Source",
			opts: &Options{
				Src:      "press",
				SrcFile:  tmpFile.Name(),
				SrcSpeed: 10,
				Dst:      "nil",
			},
			wantErr: false,
		},
		{
			name: "Invalid Press Source - Missing File",
			opts: &Options{
				Src:      "press",
				SrcSpeed: 10,
				Dst:      "nil",
			},
			wantErr: true,
		},
		{
			name: "Invalid Press Source - File Not Found",
			opts: &Options{
				Src:      "press",
				SrcFile:  "non_existent_file.txt",
				SrcSpeed: 10,
				Dst:      "nil",
			},
			wantErr: true,
		},

		// --- Destination Validations ---
		{
			name: "Valid Kafka Dest",
			opts: &Options{
				Src:                 "kafka",
				SrcBootstrapServers: []string{"localhost:9092"},
				SrcGroup:            "group1",
				SrcTopic:            "topic1",
				Dst:                 "kafka",
				DstBootstrapServers: []string{"localhost:9092"},
				DstTopic:            "topic-out",
			},
			wantErr: false,
		},
		{
			name: "Invalid Kafka Dest - Missing Topic",
			opts: &Options{
				Src:                 "kafka",
				SrcBootstrapServers: []string{"localhost:9092"},
				SrcGroup:            "group1",
				SrcTopic:            "topic1",
				Dst:                 "kafka",
				DstBootstrapServers: []string{"localhost:9092"},
			},
			wantErr: true,
		},
		{
			name: "Valid NSQ Dest",
			opts: &Options{
				Src:                 "kafka",
				SrcBootstrapServers: []string{"localhost:9092"},
				SrcGroup:            "group1",
				SrcTopic:            "topic1",
				Dst:                 "nsq",
				DstNSQDTCPAddresses: []string{"localhost:4150"},
				DstTopic:            "topic-out",
			},
			wantErr: false,
		},
		{
			name: "Valid File Dest",
			opts: &Options{
				Src:                 "kafka",
				SrcBootstrapServers: []string{"localhost:9092"},
				SrcGroup:            "group1",
				SrcTopic:            "topic1",
				Dst:                 "file",
				DstFile:             "non_existent_output.txt", // Should not exist
				DstBufSizeByte:      1024,
			},
			wantErr: false,
		},
		{
			name: "Invalid File Dest - File Exists",
			opts: &Options{
				Src:                 "kafka",
				SrcBootstrapServers: []string{"localhost:9092"},
				SrcGroup:            "group1",
				SrcTopic:            "topic1",
				Dst:                 "file",
				DstFile:             existingDstFile.Name(), // Exists, should fail
				DstBufSizeByte:      1024,
			},
			wantErr: true,
		},
		{
			name: "Invalid Chan Size",
			opts: &Options{
				Src:                 "kafka",
				SrcBootstrapServers: []string{"localhost:9092"},
				SrcGroup:            "group1",
				SrcTopic:            "topic1",
				Dst:                 "nil",
				ChanSize:            -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.opts.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Options.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}