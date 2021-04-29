package wu2quant

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestNoDeps(t *testing.T) {
	if os.Getenv("WU2QUANT_SKIP_MOD") != "" {
		// Use this to avoid this check if you need to use spew.Dump in tests:
		t.Skip()
	}

	fix, err := ioutil.ReadFile("go.mod.fix")
	if err != nil {
		panic(err)
	}

	{
		bts, err := ioutil.ReadFile("go.mod")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(fix, bts) {
			t.Fatal("go.mod contains unexpected content")
		}
	}

	{
		_, err := ioutil.ReadFile("go.sum")
		if !os.IsNotExist(err) {
			t.Fatal("go.sum should not exist")
		}
	}
}
