# -*- coding: utf-8 -*-
from __future__ import print_function


def verify_result(result):
    pass_checks = ['OutputPass', 'TestsPass']
    for check in pass_checks:
        if check in result and not result[check]:
            return False
    return True
