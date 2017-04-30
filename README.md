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
    * Add directory and supporting files for the challenge in the `var/challenges` dir. A README will automatically be created in the challenge directory based on the description in challenges.yaml.
    * Add an example solution.
    * Add expected output if the command has output that should be verified.
    * Add a test script to `ro_volume/cmdtests/` if tests are needed after the command is run (see other challenge examples).
    * Add your github username or name to the author field.
* Run `make test` to make sure your new challenge works with the example.
* Submit a pull request.

## Installation
* Install docker on your machine
* `virtualenv -p $(which python2.7) venv2.7`
* `source venv2.7/bin/activate`
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
