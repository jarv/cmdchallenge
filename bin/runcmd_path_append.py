from sys import path
from os.path import dirname, realpath, join

dir_path = dirname(realpath(__file__))
path.append(join(dir_path, "../lambda_src/runcmd"))
