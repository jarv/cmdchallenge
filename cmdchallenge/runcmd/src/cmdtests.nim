import tables
import strformat
import strutils
import os
import json
import sequtils
import sugar
import oswalkdir

proc existsFileNotSymlink(fname: string): bool =
  return (existsFile(fname) and not symlinkExists(fname))

proc chCreateFile(jsonChallenge: JsonNode): (string, bool) =
  if not existsFileNotSymlink("take-the-command-challenge"):
    return (&"Test failed, file does not exist", false)
  
  if readFile("take-the-command-challenge") != "":
    return (&"Test failed, file is not empty", false)

  ("", true)

proc chCreateDirectory(jsonChallenge: JsonNode): (string, bool) =
  if not existsDir("tmp/files"):
    return (&"Test failed, directory does not exist", false)

  ("", true)

proc chCopyFile(jsonChallenge: JsonNode): (string, bool) =
  if not existsFileNotSymlink("tmp/files/take-the-command-challenge"):
    return (&"Test failed, file does not exist", false)

  if not existsFileNotSymlink("take-the-command-challenge"):
    return (&"Test failed, original file still exists", false)

  ("", true)

proc chMoveFile(jsonChallenge: JsonNode): (string, bool) =
  if not existsFileNotSymlink("tmp/files/take-the-command-challenge"):
    return (&"Test failed, file does not exist", false)

  if existsFile("take-the-command-challenge"):
    return (&"Test failed, file was not moved", false)

  if readFile("tmp/files/take-the-command-challenge") != "":
    return (&"Test failed, file was modified", false)

  ("", true)

proc chCreateSymlink(jsonChallenge: JsonNode): (string, bool) =
  if not symlinkExists("take-the-command-challenge"):
    return (&"Test failed, symlink does not exist", false)

  let fullSymlinkPath = expandFilename(expandSymlink("take-the-command-challenge"))

  if fullSymlinkPath != "/var/challenges/create_symlink/tmp/files/take-the-command-challenge":
    return (&"Test failed, symlink does not point to tmp/files/take-the-command-challenge", false)

  ("", true)

proc chDeleteFiles(jsonChallenge: JsonNode): (string, bool) =
  try:
    if not existsDir("/var/challenges/delete_files"):
      return (&"Test failed, challenge directory was removed", false)
  except IOError:
    return (&"Test failed, challenge directory was removed", false)

  let files = toSeq(walkDirRec("."))
  if files.len > 0:
    return (&"Test failed, {files.len} files or directories remain", false)

  ("", true)

proc chRemoveExtensionsFromFiles(jsonChallenge: JsonNode): (string, bool) =
  for f in walkDirRec("."):
    let (dir, name, ext) = splitFile(f)
    if ext != "":
      return (&"Test failed, there is still one or more files with an extension", false)

  ("", true)

proc chRemoveFilesWithADash(jsonChallenge: JsonNode): (string, bool) =
  if toSeq(walkDirRec(".")).filterIt(it.existsFile).len != 1:
      return (&"Test failed, expecting one file", false)

  for f in walkDirRec("."):
    if "-" in f:
      return (&"Test failed, one or more files with a dash in the name exists", false)

  ("", true)

proc chRemoveFilesWithExtension(jsonChallenge: JsonNode): (string, bool) =
  if toSeq(walkDirRec(".")).filterIt(it.existsFile).len != 4:
      return (&"Test failed, expecting 4 files", false)
  for f in walkDirRec("."):
    let (dir, name, ext) = splitFile(f)
    if ext == ".doc":
      return (&"Test failed, found file '{f}' with a .doc extension", false)

  ("", true)

proc chRemoveFilesWithoutExtension(jsonChallenge: JsonNode): (string, bool) =
  if toSeq(walkDirRec(".")).filterIt(it.existsFile).len != 4:
      return (&"Test failed, expecting 4 files", false)

  for f in walkDirRec("."):
    if not f.existsFile:
      continue
    let (_, _, ext) = splitFile(f)
    if not (ext in [".txt", ".exe"]):
      return (&"Test failed, found file '{f}' without a .txt or .exe extension", false)

  ("", true)

proc chReplaceTextInFiles(jsonChallenge: JsonNode): (string, bool) =
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
}.toTable

proc runCmdTest*(jsonChallenge: JsonNode): (string, bool) =
  let slug = jsonChallenge["slug"].getStr
  if cmdTests.hasKey(slug):
    return cmdTests[slug](jsonChallenge)
  else:
    return ("", true)

