package challenge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const helloWorldJSON = `
{
  "slug": "hello_world",
  "emoji": "emojis/1F40C",
  "disp_title": "hello world",
  "version": 5,
  "example": "echo 'hello world'",
  "expected_failures": [
    "echo \"nope\""
  ],
  "expected_output": {
    "lines": [
      "hello world"
    ]
  }
}
`

const expectedMultiOrdered = `
{
  "expected_output": {
    "lines": [
      "1",
      "2",
      "3"
    ]
  }
}`
const expectedMultiNotOrdered = `
{
  "expected_output": {
	"order": false,
    "lines": [
      "3",
      "1",
      "2"
    ]
  }
}`
const expectedMultiReSub = `
{
  "expected_output": {
    "re_sub": ["^.*\/", ""],
    "lines": [
      "file1",
      "file2",
      "file3"
    ]
  }
}`

func TestHasExpectedLines(t *testing.T) {
	assert.True(t, fakeHelloWorldCh(t).HasExpectedLines())
}

func TestExpectedOutput(t *testing.T) {
	assert.Equal(t, fakeHelloWorldCh(t).ExpectedLines(), []string{`hello world`})
}

func TestExpectedFailures(t *testing.T) {
	assert.Equal(t, fakeHelloWorldCh(t).ExpectedFailures(), []string{`echo "nope"`})
}

func TestExample(t *testing.T) {
	assert.Equal(t, fakeHelloWorldCh(t).Example(), `echo 'hello world'`)
}

func TestSlug(t *testing.T) {
	assert.Equal(t, fakeHelloWorldCh(t).Slug(), `hello_world`)
}

func TestVersion(t *testing.T) {
	assert.Equal(t, fakeHelloWorldCh(t).Version(), 5)
}

func TestDir(t *testing.T) {
	assert.Equal(t, fakeHelloWorldCh(t).Dir(), `hello_world`)
}

func TestImg(t *testing.T) {
	assert.Equal(t, fakeHelloWorldCh(t).Img(), `cmd`)
}

func TestHasOrderedExpectedLines(t *testing.T) {
	assert.True(t, fakeHelloWorldCh(t).HasOrderedExpectedLines())
}

func TestMatchesLines(t *testing.T) {
	testCases := []struct {
		name   string
		chJSON string
		cmdOut string
		want   bool
	}{
		{
			name:   "match: single line",
			chJSON: helloWorldJSON,
			cmdOut: "hello world\n",
			want:   true,
		},
		{
			name:   "no match: single line",
			chJSON: helloWorldJSON,
			cmdOut: "no match\n",
			want:   false,
		},
		{
			name:   "match: multi line, ordered",
			chJSON: expectedMultiOrdered,
			cmdOut: "1\n2\n3\n",
			want:   true,
		},
		{
			name:   "no match: multi line, ordered",
			chJSON: expectedMultiOrdered,
			cmdOut: "3\n1\n2\n",
			want:   false,
		},
		{
			name:   "match: multi line, not ordered",
			chJSON: expectedMultiNotOrdered,
			cmdOut: "3\n1\n2\n",
			want:   true,
		},
		{
			name:   "match: multi line, resub",
			chJSON: expectedMultiReSub,
			cmdOut: "/file1\n/file2\n/file3\n",
			want:   true,
		},
		{
			name:   "match: multi line, resub",
			chJSON: expectedMultiReSub,
			cmdOut: "file1\nfile2\nfile3\n",
			want:   true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ch, err := NewChallenge([]byte(tt.chJSON))
			require.NoError(t, err)
			actual, err := ch.MatchesLines(tt.cmdOut, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		})
	}
}

func fakeHelloWorldCh(t *testing.T) *Challenge {
	ch, err := NewChallenge([]byte(helloWorldJSON))
	require.NoError(t, err)

	return ch
}
