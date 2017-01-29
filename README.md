# CMD Challenge

This repo holds the challenges for [cmdchallenge.com](https://cmdchallenge.com)

```
 _____ _____ ____     _____ _____ _____ __    __    _____ _____ _____ _____ 
|     |     |    \   |     |  |  |  _  |  |  |  |  |   __|   | |   __|   __|
|   --| | | |  |  |  |   --|     |     |  |__|  |__|   __| | | |  |  |   __|
|_____|_|_|_|____/   |_____|__|__|__|__|_____|_____|_____|_|___|_____|_____|
```

## Contribute

* Add a new entry to [challenges.yml](https://github.com/jarv/cmdchallenge/blob/master/challenges.yaml).
    * Pick a unique slug name.
    * Type a description.
    * Add directory and supporting files for the challenge in the `var/challenges` dir.
    * Add an example solution.
    * Add expected output or tests for the verification (see other challenges for examples examples).
    * Add your github username or name to the author field.
* Submit a pull request

## Installation
* Install docker on your machine
* `pip install -r requirements.txt`

## Testing

* `make test`

Assuming you have docker installed running `make test` will create a new
docker image, load it and run all tests.

* `./bin/test_challenges <test_name>`
Test a single challenge using the currently built docker container or
all challenges (faster than `make test`).

## Bugs / Suggestions

* Open [a github issue](https://github.com/jarv/cmdline-challenges/issues).
