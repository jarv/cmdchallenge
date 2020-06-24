import tables
import strformat
import strutils
import os
import json
import sequtils
import sugar
import oswalkdir

proc delete_files(jsonChallenge: JsonNode): (string, bool) =
  let files = toSeq(walkDirRec("."))
  if files.len > 0:
    return (&"Test failed, {files.len} files or directories remain", false)
  ("", true)

proc remove_extensions_from_files(jsonChallenge: JsonNode): (string, bool) =
  for f in walkDirRec("."):
    let (dir, name, ext) = splitFile(f)
    if ext != "":
      return (&"Test failed, there is still one or more files with an extension", false)
  ("", true)

proc remove_files_with_a_dash(jsonChallenge: JsonNode): (string, bool) =
  if toSeq(walkDirRec(".")).filterIt(it.existsFile).len != 1:
      return (&"Test failed, expecting one file", false)

  for f in walkDirRec("."):
    if "-" in f:
      return (&"Test failed, one or more files with a dash in the name exists", false)
  ("", true)

proc remove_files_with_extension(jsonChallenge: JsonNode): (string, bool) =
  if toSeq(walkDirRec(".")).filterIt(it.existsFile).len != 4:
      return (&"Test failed, expecting 4 files", false)
  for f in walkDirRec("."):
    let (dir, name, ext) = splitFile(f)
    if ext == ".doc":
      return (&"Test failed, found file '{f}' with a .doc extension", false)
  ("", true)

proc remove_files_without_extension(jsonChallenge: JsonNode): (string, bool) =
  if toSeq(walkDirRec(".")).filterIt(it.existsFile).len != 4:
      return (&"Test failed, expecting 4 files", false)
  for f in walkDirRec("."):
    if not f.existsFile:
      continue
    let (_, _, ext) = splitFile(f)
    if not (ext in [".txt", ".exe"]):
      return (&"Test failed, found file '{f}' without a .txt or .exe extension", false)
  ("", true)

proc replace_text_in_files(jsonChallenge: JsonNode): (string, bool) =
  if toSeq(walkDirRec(".")).filterIt(it.existsFile and it.endswith(".txt")).len != 3:
    return (&"Test failed, expecting 3 .txt files", false)
  for fname in walkDirRec("."):
    if fname.existsFile and fname.endswith(".txt"):
      for line in fname.lines:
        if "challenges are difficult" in line:
          return ("Test failed, found the string 'challenges are difficult", false)
  
  if not ("challenges are difficult" in readFile("not-a-text-file")):
    return ("Test failed, files without .txt extension must remain unmodified.", false)
  ("", true)


let cmdTests = {
  "delete_files": delete_files,
  "remove_extensions_from_files": remove_extensions_from_files,
  "remove_files_with_a_dash": remove_files_with_a_dash,
  "remove_files_with_extension": remove_files_with_extension,
  "remove_files_without_extension": remove_files_without_extension,
  "replace_text_in_files": replace_text_in_files,
}.toTable

proc runCmdTest*(jsonChallenge: JsonNode): (string, bool) =
  let slug = jsonChallenge["slug"].getStr
  if cmdTests.hasKey(slug):
    return cmdTests[slug](jsonChallenge)
  else:
    return ("", true)

