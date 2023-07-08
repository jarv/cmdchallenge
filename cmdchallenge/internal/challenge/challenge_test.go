package challenge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const helloWorldYAML = `---
- slug: hello_world
  version: 5
  example: echo 'hello world'
  expected_failures:
    - echo "nope"
  expected_output:
    lines:
      - 'hello world'
`

const expectedMultiOrdered = `---
- slug: expectedMultiOrdered
  expected_output:
    lines:
      - 1 
      - 2 
      - 3
`
const expectedMultiNotOrdered = `---
- slug: expectedMultiNotOrdered
  expected_output:
    order: false
    lines:
      - 3 
      - 2 
      - 1
`
const expectedMultiReSub = `---
- slug: expectedMultiReSub
  expected_output:
    re_sub: 
      - "^.*/"
      - ""
    lines:
      - file1
      - file2
      - file3
`

const expectedRemoveNonMatching = `---
- slug: expectedRemoveNonMatching
  expected_output:
    ignore_non_matching: true
    lines:
      - single line that matches
`

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
		slug   string
		chYAML string
		cmdOut string
		want   bool
	}{
		{
			name:   "match: single line",
			slug:   "hello_world",
			chYAML: helloWorldYAML,
			cmdOut: "hello world\n",
			want:   true,
		},
		{
			name:   "no match: single line",
			slug:   "hello_world",
			chYAML: helloWorldYAML,
			cmdOut: "no match\n",
			want:   false,
		},
		{
			name:   "no match: multi line",
			slug:   "hello_world",
			chYAML: helloWorldYAML,
			cmdOut: "no match\nhello world\nno match\n",
			want:   false,
		},
		{
			name:   "match: multi line, ordered",
			slug:   "expectedMultiOrdered",
			chYAML: expectedMultiOrdered,
			cmdOut: "1\n2\n3\n",
			want:   true,
		},
		{
			name:   "no match: multi line, ordered",
			slug:   "expectedMultiOrdered",
			chYAML: expectedMultiOrdered,
			cmdOut: "3\n1\n2\n",
			want:   false,
		},
		{
			name:   "match: multi line, not ordered",
			slug:   "expectedMultiNotOrdered",
			chYAML: expectedMultiNotOrdered,
			cmdOut: "3\n1\n2\n",
			want:   true,
		},
		{
			name:   "match: multi line, resub",
			slug:   "expectedMultiReSub",
			chYAML: expectedMultiReSub,
			cmdOut: "/file1\n/file2\n/file3\n",
			want:   true,
		},
		{
			name:   "match: multi line, resub",
			slug:   "expectedMultiReSub",
			chYAML: expectedMultiReSub,
			cmdOut: "file1\nfile2\nfile3\n",
			want:   true,
		},
		{
			name:   "match: multi line, remove non-matching",
			slug:   "expectedRemoveNonMatching",
			chYAML: expectedRemoveNonMatching,
			cmdOut: "junk\nsingle line that matches\njunk\n\njunk",
			want:   true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ch, err := NewChallenge(ChallengeOptions{Slug: tt.slug, ChallengesYAML: tt.chYAML})
			require.NoError(t, err)
			actual, err := ch.MatchesLines(tt.cmdOut, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		})
	}
}

func fakeHelloWorldCh(t *testing.T) *Challenge {
	ch, err := NewChallenge(ChallengeOptions{Slug: "hello_world", ChallengesYAML: helloWorldYAML})
	require.NoError(t, err)

	return ch
}
