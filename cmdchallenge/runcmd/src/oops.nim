import osproc
import os
import strformat
import strutils

const OOPS_PROG = "oops-this-will-delete-bin-dirs"

type
   OopsProc* = object
       p*: Process
       pid*: int
       slug*: string


proc start*(oopsProc: var OopsProc, prog: string = OOPS_PROG, targetPid: int = 42) =
  if not oopsProc.slug.startsWith("oops"):
    return

  while true:
    let p = startProcess(
      command=OOPS_PROG, args=["-t", "0"], options={poUsePath}
    )
    let _ = p.waitForExit()
    p.close

    if p.processId == targetPid - 1:
      break
    
  oopsProc.p = startProcess(
    command=OOPS_PROG, options={poUsePath}
  )
  oopsProc.pid = oopsProc.p.processId

proc stop*(oopsProc: var OopsProc) =
  if not oopsProc.slug.startsWith("oops"):
    return
  if oopsProc.p != nil and oopsProc.p.running:
    oopsProc.p.terminate
    oopsProc.p.close
