package pipes

import (
	"fmt"
	"github.com/relvacode/pipe/e2e"
	"testing"
)

type ChecksumPipeTestCase struct {
	Name   string
	Using  string
	Expect string
}

func (tc ChecksumPipeTestCase) Run(t *testing.T) {
	output, err := e2e.RunConsoleTest([]byte(tc.Using), fmt.Sprintf("checksum.%s", tc.Name))
	if err != nil {
		t.Fatal(err)
	}

	if tc.Expect != output {
		t.Fatalf("Incorrect checksum, expected %q but got %q", tc.Expect, output)
	}
}

func TestChecksumPipe(t *testing.T) {
	// echo -n asdfghjkl | openssl <checksum>
	testData := "asdfghjkl"
	cases := []ChecksumPipeTestCase{
		{
			Name:   "md5",
			Using:  testData,
			Expect: "c44a471bd78cc6c2fea32b9fe028d30a",
		},
		{
			Name:   "sha1",
			Using:  testData,
			Expect: "5fa339bbbb1eeaced3b52e54f44576aaf0d77d96",
		},
		{
			Name:   "sha256",
			Using:  testData,
			Expect: "5c80565db6f29da0b01aa12522c37b32f121cbe47a861ef7f006cb22922dffa1",
		},
		{
			Name:   "sha512",
			Using:  testData,
			Expect: "9c6d7952755415c26ff3c5dc3cc1ee281b56c8542c8619e1d5133a49387d9b26deab3d0d140849a84ca8d13b34cca329af6878ab27d505ccccd473b3a7c56c2a",
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, tc.Run)
	}
}
