import osproc

const OOPS_PROG = "oops-this-will-delete-bin-dirs"

type
   OopsProc* = object
       p*: Process
       pid*: int


proc start*(oopsProc: var OopsProc, prog: string = OOPS_PROG, targetPid: int = 42) =
  while true:
    oopsProc.p = startProcess(
      command=OOPS_PROG, options={poUsePath}
    )
   
    if oopsProc.p.processId == oopsProc.pid + targetPid:
      break

    oopsProc.p.terminate

  oopsProc.pid = oopsProc.p.processId

proc stop*(oopsProc: var OopsProc) =
  oopsProc.p.terminate
