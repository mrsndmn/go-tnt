package tnt

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// iterator [0, N)
func N(n int) []struct{} {
	return make([]struct{}, n)
}

func randomInt(min int, max int) int {
	if min == max {
		return min
	}
	return rand.Intn(max-min) + min
}

// generate test values:
// * 0
// * 1
// * 2-3
// * 4-15
// ...
// * random from range [2**(k-1), 2**k-1]
// ...
// * 2**n-1
func values(n int) chan uint {
	ch := make(chan uint, n+2)

	ch <- 0
	for x := range N(n) {
		if x == 63 {
			ch <- uint(2 * randomInt(1<<uint(x-1), (1<<(uint(x-1)+1))))
		} else {
			ch <- uint(randomInt(1<<uint(x), (1 << (uint(x) + 1))))
		}
		// pp.Println(x)
	}
	ch <- 1<<uint(n) - 1
	close(ch)
	return ch
}

func python(code string) []byte {
	cmd := exec.Command("python", "-c", code)
	out, err := cmd.Output()

	if err != nil {
		log.Fatal(err.Error())
		return []byte{}
	}

	return []byte(out)
}

func pythonIproto(code string, args ...interface{}) []byte {
	codeWithParams := fmt.Sprintf(code, args...)
	res := python(
		fmt.Sprintf(
			"import python_iproto, sys; sys.stdout.write(str(python_iproto.%s).encode(\"base64\").strip())",
			codeWithParams,
		),
	)

	data, err := base64.StdEncoding.DecodeString(string(res))
	if err != nil {
		log.Fatal("error:", err)
		return []byte{}
	}

	return data
}

func TestValues(t *testing.T) {
	assert := assert.New(t)

	index := 0
	for value := range values(8) {
		switch index {
		case 0:
			assert.Equal(uint(0), value)
		case 1:
			assert.Equal(uint(1), value)
		case 9:
			assert.Equal(uint(255), value)
		default:
			assert.True(value >= (1 << uint(index-1)))
			assert.True(value < (1 << uint(index)))
		}

		index++
	}

	assert.Equal(10, index)
}

func TestPython(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]byte("BAEAAAA="),
		python("import python_iproto, sys; sys.stdout.write(str(python_iproto.pack_int(1)).encode(\"base64\").strip())"),
	)
}

func TestPythonIproto(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(
		[]byte{0x04, 0x01, 0x00, 0x00, 0x00},
		pythonIproto("pack_int(1)"),
	)
}

func TestPackB(t *testing.T) {
	assert := assert.New(t)

	for value := range values(8) {
		assert.Equal(
			pythonIproto("struct_B.pack(%d)", value),
			PackB(uint8(value)),
		)
	}
}

func TestPackL(t *testing.T) {
	assert := assert.New(t)

	for value := range values(32) {
		assert.Equal(
			pythonIproto("struct_L.pack(%d)", value),
			PackL(uint32(value)),
		)
	}
}

func TestPackQ(t *testing.T) {
	assert := assert.New(t)

	for value := range values(64) {
		assert.Equal(
			pythonIproto("struct_Q.pack(%d)", value),
			PackQ(uint64(value)),
		)
	}
}
