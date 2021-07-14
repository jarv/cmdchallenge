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
      "Error": message
    }
  echo j
  quit(0)

proc runCombinedOutput(command: string, timeout: int): (string, int) =
  let args = ["-O", "globstar", "-c", &"export MANPAGER=cat;{command}"]
  let process = startProcess(command = "bash", args = args, options = {
      poStdErrToStdOut, poUsePath})
  let ret = waitForExit(p = process, timeout = timeout)
  let strm = outputStream(p = process)
  var outp = strm.readAll
  outp.stripLineEnd
  close(p = process)
  return (outp, ret)


proc hasExpectedLines(jsonChallenge: JsonNode): bool =
  if not jsonChallenge.hasKey("expected_output"):
    return false

  return jsonChallenge["expected_output"].hasKey("lines")

proc matchesOutput(output: string, jsonChallenge: JsonNode, expectedLines: seq[
    string] = @[]): bool =
  if not hasExpectedLines(jsonChallenge):
    raise newException(ValueError, "Getting expected lines on challenge with nothing expected!")

  var cmdLines = output.splitLines

  # If expectedLines is not passed then default to the values
  # provided by the challenge
  var expected = if expectedLines.len == 0:
                   jsonChallenge["expected_output"]["lines"].getElems.mapIt(it.getStr)
                 else:
                   expectedLines

  if cmdLines.len != expected.len:
    return false

  if jsonChallenge["expected_output"].hasKey("re_sub"):
    let reSub = jsonChallenge["expected_output"]["re_sub"].getElems
    apply(cmdLines, proc (line: var string) =
      line = line.replace(re(reSub[0].getStr), by = reSub[1].getStr))

  if not jsonChallenge["expected_output"]{"order"}.getBool(true):
    return expected.sorted == cmdLines.sorted

  return expected == cmdLines

proc run(command: string, jsonChallenge: JsonNode): JsonNode =
  # For all commands, set a default timeout of 5 seconds
  let challengeTimeout = jsonChallenge{"timeout"}.getInt(5000)

  var resp = %* {}

  # For some challenges, start the oops process
  var oopsProc = OopsProc()
  if jsonChallenge["slug"].getStr.startsWith("oops"):
    oopsProc.start()

  let (output, exitCode) = runCombinedOutput(command, challengeTimeout)
  resp["Output"] = %* output
  resp["ExitCode"] = %* exitCode

  # By default, set pass to true
  resp["Correct"] = %* true

  # Check if the output matches expected lines
  if hasExpectedLines(jsonChallenge):
    let matchesOutput = matchesOutput(output, jsonChallenge)
    if matchesOutput:
      resp["OutputPass"] = %* true
    else:
      resp["OutputPass"] = %* false
      resp["Correct"] = %* false
      return resp

  # Run tests if tests are specified
  if hasTest(jsonChallenge):
    let testError = runCmdTest(jsonChallenge, oopsProc)
    if testError == "":
      resp["TestPass"] = %* true
    else:
      resp["TestPass"] = %* false
      resp["Error"] = %* testError
      resp["Correct"] = %* false
      return resp

  # Oops proc is not currently needed
  # for any of the randomizers
  if jsonChallenge["slug"].getStr.startsWith("oops"):
    oopsProc.stop()

  # If randomizers are defined, them
  if hasRandomizer(jsonChallenge):
    let expectedAfterRandomizer = runRandomizer(jsonChallenge)
    # Discard the exit code of the command when we run randomizer
    let (afterRandOutput, _) = runCombinedOutput(command, challengeTimeout)

    # Check for expected lines after randomizer
    if hasExpectedLines(jsonChallenge):
      let matchesAfterRandOutput = matchesOutput(afterRandOutput,jsonChallenge, expectedAfterRandomizer)
      if matchesAfterRandOutput:
        resp["AfterRandOutputPass"] = %* true
      else:
        resp["AfterRandOutputPass"] = %* false
        resp["Correct"] = %* false
        return resp

    # Run tests after randomizer
    if hasTest(jsonChallenge):
      let afterRandTestError = runCmdTest(jsonChallenge, oopsProc)
      if afterRandTestError == "":
        resp["AfterRandTestPass"] = %* true
      else:
        resp["AfterRandTestPass"] = %* false
        resp["Error"] = %* afterRandTestError
        resp["Correct"] = %* false
        return resp

  resp

## MAIN

var command, challenge: string
var jsonChallenge: JsonNode

let p = newParser("runcmd"):
  option("-s", "--slug", help = "slug", default = "hello_world")
  arg("cmd", default = """echo -e "./access.log\naccess.log.2\naccess.log.1"""")

var opts = p.parse

# if the base64 decode fails assume the command was passed in
# without encoding
try:
  command = decode(opts.cmd)
except ValueError:
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

try:
  echo run(command, jsonChallenge)
except:
  errorExit(getCurrentExceptionMsg())
