#!/usr/bin/env python

import getopt
import sys
import re
import os
import requests

def main():
  def usage():
    sys.stderr.write("Usage %s: \n" % sys.argv[0])
    sys.stderr.write("  -h           -- this message\n")
    sys.stderr.write("  -i <path>    -- input file\n")
    sys.stderr.write("  -o <path>    -- output file path\n")

  output_fn = None
  input_fn = None

  yfile_re = re.compile("^([ ]+)?<\?yfile include=\"(?P<fn>.*)\"([ ]+)?\?>")

  opts,remaining = getopt.getopt(sys.argv[1:], 'hi:o:', ['help', 'input=',
                                                        'output='])

  for opt, arg in opts:
    if opt in ('-o', '--output'):
      output_fn = arg
    elif opt in ('-i', '--input'):
      input_fn = arg

  if output_fn is None or input_fn is None:
    usage()
    sys.stderr.write("\n")
    sys.stderr.write("FATAL: must specify output filename\n")
    sys.exit(1)

  try:
    infh = open(input_fn, 'r')
  except IOError, m:
    sys.stderr.write("FATAL: could not open input filename (%s)" % m)
    sys.exit(1)

  if not os.path.isabs(input_fn):
    inabsdir = os.path.dirname(os.path.abspath(input_fn))
  else:
    inabsdir = os.path.dirname(input_fn)

  try:
    outfh = open(output_fn, 'w')
  except IOError, m:
    sys.stderr.write("FATAL: could not open output filename (%s)" % m)
    sys.exit(1)

  for l in infh.readlines():
    if not yfile_re.match(l):
      outfh.write(l)
    else:
      ifn = yfile_re.sub('\g<fn>', l).rstrip("\n")
      process = False
      c = None
      if re.match("^http(s)?://", ifn):
        r = requests.get(ifn)
        c = r.content
      elif re.match("^file://", ifn):
        ifn = re.sub("^file://", '', ifn)
        process = True
      elif not os.path.isabs(ifn):
        ifn = inabsdir+"/"+ifn
        process = True
      else:
        process = True

      if process:
        try:
          c = open(ifn, 'r').readlines()
        except IOError, m:
          sys.stderr.write("FATAL: could not open a referenced file (%s)\n" % m)
          sys.exit(1)

      if c is not None:
        outfh.write("\n")
        for l in c:
          outfh.write(l)
        outfh.write("\n")

if __name__ == '__main__':
  main()


