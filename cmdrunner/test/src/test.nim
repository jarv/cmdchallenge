import algorithm
import argparse
import base64
import json
import os
import osproc
import re
import sequtils
import strformat
import strutils
import oops

const OOPS_PROG = "oops-this-will-delete-bin-dirs"

while true:
  let p = startProcess(
    command=OOPS_PROG, args=["-t", "0"], options={poUsePath}
  )
  let pid = p.processId
  echo &"Staring process with pid: {pid}"
  let ret = p.waitForExit()
  p.close

  if pid == 41:
    break

while existsDir(&"/proc/41"):
  # Wait for old pid dirs to be cleaned up
  sleep(10)

let p = startProcess(
  command=OOPS_PROG, options={poUsePath}
)
