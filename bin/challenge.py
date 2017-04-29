# -*- coding: utf-8 -*-
from __future__ import print_function


def bool_to_int_dyn(value):
    if value:
        return 1
    else:
        return 2


def verify_result(result):
    pass_checks = ['OutputPass', 'TestsPass', 'AfterRandOutputPass', 'AfterRandTestsPass']
    for check in pass_checks:
        if check in result and not result[check]:
            return False
    return True


def value_rand_pass(result):
    return result.get('AfterRandOutputPass', True) and result.get('AfterRandTestsPass', True)


def value_output_rand_result(result):
    if 'AfterRandOutputPass' not in result:
        return 0
    return bool_to_int_dyn(result['AfterRandOutputPass'])


def value_tests_rand_result(result):
    if 'AfterRandTestsPass' not in result:
        return 0
    return bool_to_int_dyn(result['AfterRandTestsPass'])
