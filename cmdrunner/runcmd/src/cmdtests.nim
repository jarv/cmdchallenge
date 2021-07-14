import algorithm
import json
import oops
import os
import osproc
import sequtils
import strutils
import tables
import strformat

proc fileExistsNotSymlink(fname: string): bool =
  return (fileExists(fname) and not symlinkExists(fname))

proc chOopsKillAProcess(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  # Wait up to 500 ms for the process to be killed
  for i in toSeq(1 .. 5):
    if oopsProc.p == nil or not oopsProc.p.running:
      return ""
    sleep(10)

  "Test failed, process is still running"

proc chCreateFile(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if not fileExistsNotSymlink("take-the-command-challenge"):
    return "Test failed, file does not exist"

  elif readFile("take-the-command-challenge") != "":
    return "Test failed, file is not empty"

  ""

proc chCreateDirectory(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if not dirExists("tmp/files"):
    return "Test failed, directory does not exist"

  ""

proc chCopyFile(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if not fileExistsNotSymlink("tmp/files/take-the-command-challenge"):
    return "Test failed, file does not exist"

  if not fileExistsNotSymlink("take-the-command-challenge"):
    return "Test failed, original file was removed"

  ""

proc chMoveFile(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if not fileExistsNotSymlink("tmp/files/take-the-command-challenge"):
    return "Test failed, file does not exist"

  if fileExists("take-the-command-challenge"):
    return "Test failed, file was not moved"

  if readFile("tmp/files/take-the-command-challenge") != "":
    return "Test failed, file was modified"

  ""

proc chCreateSymlink(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if not symlinkExists("take-the-command-challenge"):
    return "Test failed, symlink does not exist"

  let fullSymlinkPath = expandFilename(expandSymlink("take-the-command-challenge"))

  if fullSymlinkPath != "/var/challenges/create_symlink/tmp/files/take-the-command-challenge":
    return "Test failed, symlink does not point to tmp/files/take-the-command-challenge"

  ""

proc chDeleteFiles(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  try:
    if not dirExists("/var/challenges/delete_files"):
      return "Test failed, challenge directory was removed"
  except IOError:
    return "Test failed, challenge directory was removed"

  let files = toSeq(walkDirRec("."))
  if files.len > 0:
    return "Test failed, {files.len} files or directories remain"

  ""

proc chRemoveExtensionsFromFiles(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  for f in walkDirRec("."):
    let (_, _, ext) = splitFile(f)
    if ext != "":
      return "Test failed, there is still one or more files with an extension"

  ""

proc chRemoveFilesWithADash(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if toSeq(walkDirRec(".")).filterIt(it.fileExists).len != 1:
    return "Test failed, expecting one file"

  for f in walkDirRec("."):
    if "-" in f:
      return "Test failed, one or more files with a dash in the name exists"

  ""

proc chRemoveFilesWithExtension(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if toSeq(walkDirRec(".")).filterIt(it.fileExists).len != 4:
    return "Test failed, expecting 4 files"
  for f in walkDirRec("."):
    let (_, _, ext) = splitFile(f)
    if ext == ".doc":
      return "Test failed, found file '{f}' with a .doc extension"

  ""

proc chRemoveFilesWithoutExtension(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if toSeq(walkDirRec(".")).filterIt(it.fileExists).len != 4:
    return "Test failed, expecting 4 files"

  for f in walkDirRec("."):
    if not f.fileExists:
      continue
    let (_, _, ext) = splitFile(f)
    if not (ext in [".txt", ".exe"]):
      return "Test failed, found file '{f}' without a .txt or .exe extension"

  ""

proc chReplaceTextInFiles(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  if toSeq(walkDirRec(".")).filterIt(it.fileExists and it.endswith(".txt")).len != 3:
    return "Test failed, expecting 3 .txt files"

  for fname in walkDirRec("."):
    if fname.fileExists and fname.endswith(".txt"):
      for line in fname.lines:
        if "challenges are difficult" in line:
          return "Test failed, found the string 'challenges are difficult"

  if not ("challenges are difficult" in readFile("not-a-text-file")):
    return "Test failed, files without .txt extension must remain unmodified."

  ""

proc ch12Days8(jsonChallenge: JsonNode, oopsProc: OopsProc): (string) =
  let elves = toSeq(walkDirRec("Elves"))
  if elves.len != 0:
    return "Test failed, elves are still in Elves/"
  let workshop = toSeq(walkDirRec("Workshop")).sorted

  if workshop != @["Workshop/Alabaster Snowball", "Workshop/Buddy",
      "Workshop/Bushy Evergreen", "Workshop/Hermey", "Workshop/Pepper Minstix",
      "Workshop/Shinny Upatree", "Workshop/Sugarplum Mary",
      "Workshop/Wunorse Openslae"]:
    return "Test failed, Elves are not in the Workshop!"

  ""

let cmdTests = {
  "delete_files": chDeleteFiles,
  "remove_extensions_from_files": chRemoveExtensionsFromFiles,
  "remove_files_with_a_dash": chRemoveFilesWithADash,
  "remove_files_with_extension": chRemoveFilesWithExtension,
  "remove_files_without_extension": chRemoveFilesWithoutExtension,
  "replace_text_in_files": chReplaceTextInFiles,
  "create_file": chCreateFile,
  "create_directory": chCreateDirectory,
  "create_symlink": chCreateSymlink,
  "copy_file": chCopyFile,
  "move_file": chMoveFile,
  "oops_kill_a_process": chOopsKillAProcess,
  "12days_8": ch12Days8,
}.toTable

proc hasTest*(jsonChallenge: JsonNode): bool =
  let slug = jsonChallenge["slug"].getStr
  return cmdTests.hasKey(slug)

proc runCmdTest*(jsonChallenge: JsonNode, oopsProc: OopsProc): string =
  let slug = jsonChallenge["slug"].getStr
  if not hasTest(jsonChallenge):
    raise newException(ValueError, &"Invalid test! challenge: {slug}")

  return cmdTests[slug](jsonChallenge, oopsProc)
