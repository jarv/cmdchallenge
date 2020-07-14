import tables
import random
import strformat
import strutils
import os
import json
import sequtils
import sugar

proc count_files(jsonChallenge: JsonNode): seq[string] =
  # Create a random number of empty files
  let expectedFiles = parseInt(
    jsonChallenge["expected_output"]["lines"].getElems[0].getStr
  )
  let randNum = rand(100)
  var randFnames = toSeq(1 .. randNum).mapIt(&"{it}-{rand(1000)}")

  for fname in randFnames:
    writeFile(fname, "")

  return @[&"{expectedFiles + randFnames.len}"]

proc count_string_in_line(jsonChallenge: JsonNode): seq[string] =
  let numMatches = parseInt(
    jsonChallenge["expected_output"]["lines"].getElems[0].getStr
  )

  let randNum = rand(200)

  for _ in 1 .. randNum:
    let f = open("access.log", fmAppend)
    f.writeLine("GET")
    f.close

  return @[&"{randNum + numMatches}"]

proc dirs_containing_files_with_extension(jsonChallenge: JsonNode): seq[string] =
  # Make some random files
  let expectedLines = jsonChallenge["expected_output"]["lines"].getElems.mapIt(it.getStr)

  var randFnames = toSeq(1 .. rand(10)).mapIt(&"some/random/dir/{it}-{rand(1000)}/some-file.tf")
  var randDirs: seq[string]
  
  for fname in randFnames:
    let (dir, _, _) = splitFile(fname)
    createDir(dir)
    randDirs.add(dir)
    writeFile(fname, "")

  return expectedLines.concat(randDirs)

proc find_primes(jsonChallenge: JsonNode): seq[string] =
  let expectedPrimes = parseInt(
    jsonChallenge["expected_output"]["lines"].getElems[0].getStr
  )

  # these prime numbers do not appear in random-numbers.txt
  let primes = @[2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199]

  let randNum = rand(primes.len)
  for prime in primes[0 ..< randNum]:
    let f = open("random-numbers.txt", fmAppend)
    f.writeLine(prime)
    f.close

  return @[&"{expectedPrimes + randNum}"]

proc find_tabs_in_a_file(jsonChallenge: JsonNode): seq[string] =
  let expectedTabs = parseInt(
    jsonChallenge["expected_output"]["lines"].getElems[0].getStr
  )
  let randNum = 10 # rand(20)

  for _ in 1 .. randNum:
    let f = open("file-with-tabs.txt", fmAppend)
    f.writeLine("\t")
    f.close

  return @[&"{expectedTabs + randNum}"]

proc list_files(jsonChallenge: JsonNode): seq[string] =
  let expectedFiles = jsonChallenge["expected_output"]["lines"].getElems.mapIt(it.getStr)
  let randFnames = toSeq(1 .. rand(10..20)).mapIt(&"{it}-{rand(1000)}")

  for fname in randFnames:
    writeFile(fname, "")

  return expectedFiles.concat(randFnames)

proc nested_dirs(jsonChallenge: JsonNode): seq[string] =
  let randNum = rand(1000)
  let f = open(".../  /. .the flag.txt", fmWrite)
  f.writeLine(&"{randNum}")
  f.close
  return @[&"{randNum}"]

proc sum_all_numbers(jsonChallenge: JsonNode): seq[string] =
  let expectedSum = parseInt(
    jsonChallenge["expected_output"]["lines"].getElems[0].getStr
  )
  let randNum = rand(1000)
  let f = open("sum-me.txt", fmAppend)
  f.writeLine(&"{randNum}")
  f.close

  return @[&"{expectedSum + randNum}"]

proc search_for_files_containing_string(jsonChallenge: JsonNode): seq[string] =
  let expectedFiles = jsonChallenge["expected_output"]["lines"].getElems.mapIt(it.getStr)
  let randFnames = toSeq(5 .. rand(13..20)).mapIt(&"access.log.{it}")

  for fname in randFnames:
    writeFile(fname, "500")

  return expectedFiles.concat(randFnames)

let randomizers = {
  "count_files": count_files,
  "count_string_in_line": count_string_in_line,
  "dirs_containing_files_with_extension": dirs_containing_files_with_extension,
  "find_primes": find_primes,
  "find_tabs_in_a_file": find_tabs_in_a_file,
  "list_files": list_files,
  "nested_dirs": nested_dirs,
  "sum_all_numbers": sum_all_numbers,
  "search_for_files_containing_string": search_for_files_containing_string,
}.toTable

proc runRandomizer*(jsonChallenge: JsonNode): seq[string] =
  randomize()
  let slug = jsonChallenge["slug"].getStr
  if randomizers.hasKey(slug):
    return randomizers[slug](jsonChallenge)
