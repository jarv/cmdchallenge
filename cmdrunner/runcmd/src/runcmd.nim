import algorithm
import argparse
import base64
import json
import os
import osproc
import randomizers
import cmdtests
import re
import sequtils
import strformat
import strutils
import oops

proc errorExit(message: string): void =
  var j = %*
    {
      "Error":  message
    }
  echo j
  quit(0)

proc runCombinedOutput(command: string, timeout: int): (string, int) =
  let args = ["-O", "globstar", "-c", &"export MANPAGER=cat;{command}"]
  let process = startProcess(command="bash", args=args, options={poStdErrToStdOut, poUsePath})
  let ret = waitForExit(p=process, timeout=timeout)
  let strm = outputStream(p=process)
  var outp = strm.readAll
  outp.stripLineEnd
  close(p=process)
  return (outp, ret)


proc matchesOutput(cmdOut: string, jsonChallenge: JsonNode, expectedLines: seq[string] = @[]): bool =
  if not jsonChallenge.hasKey("expected_output"):
    return true

  let expectedOutput = jsonChallenge["expected_output"]

  if not expectedOutput.hasKey("lines"):
    return true


  var cmdLines = cmdOut.splitLines

  # If expectedLines is not passed then default to the values
  # provided by the challenge
  var expectedLines = if expectedLines.len > 0:
                        expectedLines
                      else:
                        expectedOutput["lines"].getElems.mapIt(it.getStr)

  if cmdLines.len != expectedLines.len:
    return false

  if expectedOutput.hasKey("re_sub"):
    let reSub = expectedOutput["re_sub"].getElems
    apply(cmdLines, proc (line: var string) =
      line = line.replace(re(reSub[0].getStr), by=reSub[1].getStr))

  let orderMatters = expectedOutput{"order"}.getBool(true)

  if not orderMatters:
    expectedLines.sort
    cmdLines.sort

  return expectedLines == cmdLines

## MAIN

var command,challenge :string
var jsonChallenge :JsonNode

let p = newParser("runcmd"):
  option("-s", "--slug", help="slug", default="hello_world")
  arg("cmd", default="""echo -e "./access.log\naccess.log.2\naccess.log.1"""")

var opts = p.parse

# if the base64 decode fails assume the command was passed in
# without encoding
try:
  command = decode(opts.cmd)
except ValueError as e:
  command = opts.cmd

let progDir = getAppDir()
let challengeFname = &"{progDir}/ch/{opts.slug}.json"

try:
  challenge = readFile(challengeFname)
except IOError:
  errorExit(&"Unable to find challenge '{challengeFname}'")

try:
  jsonChallenge = parseJson(challenge)
except JsonParsingError:
  errorExit(&"Unable to parse challenge '{opts.slug}.json'")

try:
  assert jsonChallenge{"slug"}.getStr == opts.slug
except AssertionDefect:
  errorExit(&"'{challengeFname}' has an incorrect slug")

# For all commands, set a default timeout of 5 seconds
let challengeTimeout = jsonChallenge{"timeout"}.getInt(5000)


var
  outputPass, testsPass, afterRandOutputPass, afterRandTestsPass: bool = true
  cmdExitCode, afterRandExitCode: int = 0
  cmdOut, testsOut, afterRandExpectedOutput, afterRandOutput, afterRandTestsOut: string

# For some challenges, start the oops process
var oopsProc = OopsProc(slug: opts.slug)
oopsProc.start()
 
(cmdOut, cmdExitCode) = runCombinedOutput(command, challengeTimeout)
outputPass = matchesOutput(cmdOut, jsonChallenge)

(testsOut, testsPass) = runCmdTest(jsonChallenge, oopsProc)

oopsProc.stop()

let expectedAfterRandomizer = runRandomizer(jsonChallenge)

if expectedAfterRandomizer.len > 0:
  (afterRandOutput, afterRandExitCode) = runCombinedOutput(command, challengeTimeout)
  afterRandOutputPass = matchesOutput(afterRandOutput, jsonChallenge, expectedAfterRandomizer)
  (afterRandTestsOut, afterRandTestsPass) = runCmdTest(jsonChallenge, oopsProc)

var j = %*
  {
    "CmdOut":  cmdOut,
    "CmdExitCode": cmdExitCode,
    "OutputPass": outputPass,
    "TestsPass": testsPass,
    "TestsOut": testsOut,
    "AfterRandOutputPass": afterRandOutputPass,
    "AfterRandExpectedOutput": join(expectedAfterRandomizer, "\n"),
    "AfterRandOutput": afterRandOutput,
    "AfterRandTestsPass": afterRandTestsPass,
    "AfterRandTestsOut": afterRandTestsOut
  }

echo j
