package main

import (
	"testing"
)

func checkTestErr(err error, t *testing.T) {
	if err != nil {
		t.Error("Expected no error, got ", err)
	}
}

func TestNoOuput(t *testing.T) {
	var c3Blob = []byte(`{"slug": "c3", "version": 4, "author": "cmdchallenge", "description": "foo", "example": "bar"}`)
	c3Challenge, err := readChallenge(c3Blob)
	checkTestErr(err, t)
	if c3Challenge.HasExpectedOutput() != false {
		t.Error("Expected HasOutput to be false, got ", c3Challenge.HasExpectedOutput())
	}
	m1 := c3Challenge.MatchesOutput("hello\nworld\n")
	if m1 != true {
		t.Error("Expected match to be true, got ", m1)
	}
}

func TestOuputWithRegex(t *testing.T) {
	var c2Blob = []byte(`{"slug": "c2", "author": "cmdchallenge", "description": "foo", "example": "bar", "expected_output":
							{"order": false, "re_sub": ["^\\./", ""], "lines": ["hello", "world"]}}`)
	c2Challenge, err := readChallenge(c2Blob)
	checkTestErr(err, t)
	m1 := c2Challenge.MatchesOutput("hello\nworld\n")
	if m1 != true {
		t.Error("Expected match to be true, got ", m1)
	}
	m2 := c2Challenge.MatchesOutput("./hello\n./world\n")
	if m2 != true {
		t.Error("Expected match to be true, got ", m2)
	}

	m3 := c2Challenge.MatchesOutput("../hello\n./world\n")
	if m3 != false {
		t.Error("Expected match to be false, got ", m3)
	}
}

func TestOuputNoOrder(t *testing.T) {
	var c1Blob = []byte(`{"slug": "c1", "version": 4, "author": "cmdchallenge", "description": "foo", "example": "bar",
		"expected_output": {"order": false, "lines": ["herp", "derp"]}}`)
	c1Challenge, err := readChallenge(c1Blob)
	checkTestErr(err, t)
	m1 := c1Challenge.MatchesOutput("herp\nderp\n")
	if m1 != true {
		t.Error("Expected match to be true, got ", m1)
	}
	m2 := c1Challenge.MatchesOutput("derp\nherp\n")
	if m2 != true {
		t.Error("Expected match to be true, got ", m2)
	}
}

func TestOuputOrder(t *testing.T) {
	var c1Blob = []byte(`{"slug": "c1", "version": 4, "author": "cmdchallenge", "description": "foo", "example": "bar",
		 "expected_output": {"lines": ["herp", "derp"]}}`)
	c1Challenge, err := readChallenge(c1Blob)
	checkTestErr(err, t)
	mTrue := c1Challenge.MatchesOutput("herp\nderp\n")
	if mTrue != true {
		t.Error("Expected match to be true, got ", mTrue)
	}
	mFalse := c1Challenge.MatchesOutput("derp\nherp\n")
	if mFalse != false {
		t.Error("Expected match to be false, got ", mFalse)
	}
}

func TestReadChallenges(t *testing.T) {
	var c1Blob = []byte(`{"slug": "c1", "version": 4, "author": "cmdchallenge", "description": "foo", "example": "bar",
		 "expected_output": {"lines": ["herp", "derp"]}}`)
	var c2Blob = []byte(`{"slug": "c2", "author": "cmdchallenge", "description": "foo", "example": "bar", "expected_output":
	{"order": false, "re_sub": ["^\\./", ""], "lines": ["hello world"]}}`)
	var c3Blob = []byte(`{"slug": "c3", "version": 4, "author": "cmdchallenge", "description": "foo", "example": "bar"}`)
	c1Challenge, err := readChallenge(c1Blob)
	checkTestErr(err, t)
	c2Challenge, err := readChallenge(c2Blob)
	checkTestErr(err, t)
	c3Challenge, err := readChallenge(c3Blob)
	checkTestErr(err, t)
	if c2Challenge.Slug != "c2" {
		t.Error("Invalid slug name", c2Challenge.Slug)
	}
	if c3Challenge.Slug != "c3" {
		t.Error("Invalid slug name", c2Challenge.Slug)
	}
	if c1Challenge.ExpectedOutput.Order != true {
		t.Error("For order expected true but got", c1Challenge.ExpectedOutput.Order)
	}
}
