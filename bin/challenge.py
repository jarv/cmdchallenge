# -*- coding: utf-8 -*-
from __future__ import print_function
import re
from itertools import izip_longest


def cmp_lines(a, b):
    return a == b


def cmp_lines_regex(a, b):
    return re.match(a, b)


def verify_output(challenge, output, testing=False):
    if 'expected_output' in challenge:
        if not check_expected_output(output, challenge['expected_output'], testing):
            return False
    return True


def check_expected_output(output, expected_output, testing):
    output_lines = output.split('\n')
    lines = [l.encode('utf-8') if isinstance(l, basestring) else str(l) for l in expected_output['lines']]

    if 'regex' in expected_output and expected_output['regex']:
        cmp_func = cmp_lines_regex
    else:
        cmp_func = cmp_lines

    if 're_sub' in expected_output:
        output_lines = [re.sub(expected_output['re_sub'][0], expected_output['re_sub'][1], l) for l in output_lines]

    if 'order' in expected_output and expected_output['order'] is False:
        lines = sorted(lines)
        output_lines = sorted(output_lines)
    for a, b in izip_longest(lines, output_lines):
        if cmp_func(a, b):
            if testing:
                print(u"\x1b[1;32;40m\u2713\x1b[0m", end="")
        else:
            if testing:
                print(u"\x1b[1;31;40m\u2718\x1b[0m", end="")
            return(False)
    return True
