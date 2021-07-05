import os
import argparse
import sequtils

let p = newParser("oops"):
  option("-t", "--timeout", help="how many seconds to sleep", default="10000")

var opts = p.parse

let timeout = parseInt(opts.timeout)
for _ in repeat(0, timeout):
  os.sleep(1000)
